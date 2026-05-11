package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/service/order/mocks"
)

func TestPay(t *testing.T) {
	t.Parallel()

	type args struct {
		req model.PayOrderRq
	}

	type expected struct {
		res      uuid.UUID
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx    = context.Background()
		id     = uuid.New()
		tranId = uuid.New()
		num    = gofakeit.Int64()
		date   = time.Now()

		req = model.PayOrderRq{
			OrderUUID:     id,
			PaymentMethod: model.PaymentMethodCard,
		}
		items = []model.OrderItem{
			{Uuid: id, Type: model.PartTypeHull},
			{Uuid: id, Type: model.PartTypeEngine},
		}

		modelToGet = model.Order{
			OrderUUID:       id,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPendingPayment,
			CreatedAt:       date,
			UpdatedAt:       &date,
		}
		modelCancelled = model.Order{
			OrderUUID:       id,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusCancelled,
			CreatedAt:       date,
			UpdatedAt:       &date,
		}
		modelPaid = model.Order{
			OrderUUID:       id,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPaid,
			CreatedAt:       date,
			UpdatedAt:       &date,
		}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient)
		expected  expected
	}{
		{
			name: "успешная оплата - ошибок нет",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelToGet, nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return !o.UpdatedAt.IsZero() &&
						o.Status == model.OrderStatusPaid &&
						o.PaymentMethod != nil &&
						o.TransactionUUID != nil
				})).Return(nil)
				p.EXPECT().PayOrder(ctx, req).Return(tranId, nil)
			},
			expected: expected{
				res: tranId,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderConflict",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(model.Order{}, errs.ErrOrderConflict)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderConflict))
				},
			},
		},
		{
			name: "заказ уже оплачен - errs.ErrOrderAlreadyPaid",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelPaid, nil)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderAlreadyPaid))
				},
			},
		},
		{
			name: "заказ отменен - errs.ErrOrderAlreadyCancelled",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelCancelled, nil)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderAlreadyCancelled))
				},
			},
		},
		{
			name: "ошибка PaymentService - errInternal",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelToGet, nil)
				p.EXPECT().PayOrder(ctx, req).Return(uuid.Nil, errInternal)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelToGet, nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return !o.UpdatedAt.IsZero() &&
						o.Status == model.OrderStatusPaid &&
						o.PaymentMethod != nil &&
						o.TransactionUUID != nil
				})).Return(errInternal)
				p.EXPECT().PayOrder(ctx, req).Return(tranId, nil)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := mocks.NewOrderRepository(t)
			i := mocks.NewInventoryServiceClient(t)
			p := mocks.NewPaymentServiceClient(t)
			s := NewService(r, i, p)

			test.setupMock(r, p)

			res, err := s.Pay(ctx, test.args.req)

			test.expected.errCheck(t, err)

			require.Equal(t, test.expected.res, res)
		})
	}
}
