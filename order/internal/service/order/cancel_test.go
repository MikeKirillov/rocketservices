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

func TestCancel(t *testing.T) {
	t.Parallel()

	type args struct {
		id uuid.UUID
	}

	type expected struct {
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx   = context.Background()
		uuid  = uuid.New()
		num   = gofakeit.Int64()
		date  = time.Now()
		items = []model.OrderItem{
			{Uuid: uuid, Type: model.PartTypeHull},
			{Uuid: uuid, Type: model.PartTypeEngine},
		}

		modelToSave = model.Order{
			OrderUUID:       uuid,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPendingPayment,
			CreatedAt:       date,
			UpdatedAt:       &date,
		}

		errInternal = errors.New("внутренняя ошибка бд")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.OrderRepository)
		expected  expected
	}{
		{
			name: "успешная отмена - ошибок нет",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(ctx, uuid).Return(modelToSave, nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return o.Status == model.OrderStatusCancelled &&
						!o.UpdatedAt.IsZero()
				})).Return(nil)
			},
			expected: expected{
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderConflict",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(ctx, uuid).Return(model.Order{}, errs.ErrOrderConflict)
			},
			expected: expected{
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderConflict))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(ctx, uuid).Return(modelToSave, nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return o.Status == model.OrderStatusCancelled &&
						!o.UpdatedAt.IsZero()
				})).Return(errInternal)
			},
			expected: expected{
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

			test.setupMock(r)

			err := s.Cancel(ctx, test.args.id)

			test.expected.errCheck(t, err)
		})
	}
}
