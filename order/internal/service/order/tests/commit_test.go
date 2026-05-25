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

func TestCommit(t *testing.T) {
	t.Parallel()

	type args struct {
		id uuid.UUID
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

		items = []model.OrderItem{
			{Uuid: id, Type: model.PartTypeHull},
			{Uuid: id, Type: model.PartTypeEngine},
		}

		modelToGet = model.Order{
			OrderUUID:  id,
			Items:      items,
			TotalPrice: num,
			Status:     model.OrderStatusPendingPayment,
			CreatedAt:  date,
			UpdatedAt:  nil,
		}
		modelAssembled = model.Order{
			OrderUUID:       id,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusAssembled,
			CreatedAt:       date,
			UpdatedAt:       &date,
		}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager)
		expected  expected
	}{
		{
			name: "успешный коммит - ошибок нет",
			args: args{id: id},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(ctx, id).Return(modelToGet, nil)
				i.EXPECT().CommitParts(ctx, modelToGet).Return(nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return !o.UpdatedAt.IsZero() && o.Status == model.OrderStatusAssembled
				})).Return(nil)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(nil)
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
			args: args{id: id},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(ctx, id).Return(model.Order{}, errs.ErrOrderConflict)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errs.ErrOrderConflict)
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
			name: "заказ уже собран - errs.ErrOrderAlreadyAssembled",
			args: args{id: id},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(ctx, id).Return(modelAssembled, nil)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errs.ErrOrderAlreadyAssembled)
			},
			expected: expected{
				res: uuid.Nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOrderAlreadyAssembled))
				},
			},
		},
		{
			name: "ошибка InventoryService - errInternal",
			args: args{id: id},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(ctx, id).Return(modelToGet, nil)
				i.EXPECT().CommitParts(ctx, modelToGet).Return(errInternal)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errInternal)
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
			args: args{id: id},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient, m *mocks.TxManager) {
				r.EXPECT().GetForUpdate(ctx, id).Return(modelToGet, nil)
				i.EXPECT().CommitParts(ctx, modelToGet).Return(nil)
				r.EXPECT().Update(ctx, mock.MatchedBy(func(o model.Order) bool {
					return !o.UpdatedAt.IsZero() && o.Status == model.OrderStatusAssembled
				})).Return(errInternal)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errInternal)
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
			m := mocks.NewTxManager(t)
			o := mocks.NewOrderProducerService(t)
			s := service.NewService(r, i, p, m, o)

			test.setupMock(r, i, m)

			err := s.Commit(ctx, test.args.id)

			test.expected.errCheck(t, err)
		})
	}
}
