package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/order/internal/service/order/mocks"
)

func TestCancel(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	type expected struct {
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx  = context.Background()
		uuid = gofakeit.UUID()
		num  = gofakeit.Int64()
		date = time.Now()

		modelToSave = model.Order{
			OrderUUID:       uuid,
			HullUUID:        uuid,
			EngineUUID:      uuid,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPendingPayment,
			CreatedAt:       date,
		}
		modelToCancel = model.Order{
			OrderUUID:       uuid,
			HullUUID:        uuid,
			EngineUUID:      uuid,
			ShieldUUID:      nil,
			WeaponUUID:      nil,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusCancelled,
			CreatedAt:       date,
		}

		repoModelToCancel = converter.OrderToRepoUpdate(modelToCancel)
		errInternal       = errors.New("внутренняя ошибка бд")
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
				r.EXPECT().Update(ctx, repoModelToCancel).Return(nil)
			},
			expected: expected{
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderNotFound",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(ctx, uuid).Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expected: expected{
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderNotFound))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(ctx, uuid).Return(modelToSave, nil)
				r.EXPECT().Update(ctx, repoModelToCancel).Return(errInternal)
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
