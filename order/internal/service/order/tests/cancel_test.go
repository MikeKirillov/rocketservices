package tests

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
	service "github.com/MikeKirillov/rocketservices/order/internal/service/order"
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
		id    = uuid.New()
		num   = gofakeit.Int64()
		date  = time.Now()
		items = []model.OrderItem{
			{Uuid: id, Type: model.PartTypeHull},
			{Uuid: id, Type: model.PartTypeEngine},
		}

		modelToSave = model.Order{
			OrderUUID:       id,
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
		setupMock func(i *mocks.InventoryServiceClient, r *mocks.OrderRepository, m *mocks.TxManager)
		expected  expected
	}{
		{
			name: "успешная отмена - ошибок нет",
			args: args{id: id},
			setupMock: func(i *mocks.InventoryServiceClient, r *mocks.OrderRepository, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(mock.Anything, id).Return(modelToSave, nil)
				i.EXPECT().ReleaseParts(mock.Anything, modelToSave).Return(nil)
				r.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o model.Order) bool {
					return o.Status == model.OrderStatusCancelled &&
						!o.UpdatedAt.IsZero()
				})).Return(nil)
				m.EXPECT().
					Do(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(nil)
			},
			expected: expected{
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderConflict",
			args: args{id: id},
			setupMock: func(i *mocks.InventoryServiceClient, r *mocks.OrderRepository, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(mock.Anything, id).Return(model.Order{}, errs.ErrOrderConflict)
				m.EXPECT().
					Do(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errs.ErrOrderConflict)
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
			args: args{id: id},
			setupMock: func(i *mocks.InventoryServiceClient, r *mocks.OrderRepository, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(mock.Anything, id).Return(modelToSave, nil)
				i.EXPECT().ReleaseParts(mock.Anything, modelToSave).Return(nil)
				r.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o model.Order) bool {
					return o.Status == model.OrderStatusCancelled &&
						!o.UpdatedAt.IsZero()
				})).Return(errInternal)
				m.EXPECT().
					Do(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errInternal)
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
			m := mocks.NewTxManager(t)
			o := mocks.NewOrderProducerService(t)
			s := service.NewService(r, i, p, m, o)

			test.setupMock(i, r, m)

			err := s.Cancel(ctx, test.args.id)

			test.expected.errCheck(t, err)
		})
	}
}
