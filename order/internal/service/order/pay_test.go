package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
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
		id     = gofakeit.UUID()
		tranId = gofakeit.UUID()
		num    = gofakeit.Int64()
		date   = time.Now()

		req = model.PayOrderRq{
			OrderUUID:     id,
			PaymentMethod: model.PaymentMethodCard,
		}
		res, _ = uuid.Parse(tranId)

		modelToGet = model.Order{
			OrderUUID:       id,
			HullUUID:        id,
			EngineUUID:      id,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPendingPayment,
			CreatedAt:       date,
		}
		modelCancelled = model.Order{
			OrderUUID:       id,
			HullUUID:        id,
			EngineUUID:      id,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusCancelled,
			CreatedAt:       date,
		}
		modelPaid = model.Order{
			OrderUUID:       id,
			HullUUID:        id,
			EngineUUID:      id,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPaid,
			CreatedAt:       date,
		}
		modelToPay = model.Order{
			OrderUUID:       id,
			HullUUID:        id,
			EngineUUID:      id,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: &tranId,
			PaymentMethod:   &req.PaymentMethod,
			Status:          model.OrderStatusPaid,
			CreatedAt:       date,
		}

		repoModelToPay = converter.OrderToRepoUpdate(modelToPay)
		errInternal    = errors.New("внутренняя ошибка")
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
				r.EXPECT().Update(ctx, repoModelToPay).Return(nil)
				p.EXPECT().PayOrder(ctx, req).Return(tranId, nil)
			},
			expected: expected{
				res: res,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderNotFound",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderNotFound))
				},
			},
		},
		{
			name: "заказ уже оплачен - errs.ErrOrderAlreadyPaidOrCancalled",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelPaid, nil)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderAlreadyPaidOrCancalled))
				},
			},
		},
		{
			name: "заказ отменен - errs.ErrOrderAlreadyPaidOrCancalled",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelCancelled, nil)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderAlreadyPaidOrCancalled))
				},
			},
		},
		{
			name: "ошибка PaymentService - errInternal",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, p *mocks.PaymentServiceClient) {
				r.EXPECT().Get(ctx, id).Return(modelToGet, nil)
				p.EXPECT().PayOrder(ctx, req).Return("", errInternal)
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
				r.EXPECT().Update(ctx, repoModelToPay).Return(errInternal)
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
