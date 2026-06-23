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

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id uuid.UUID
	}

	type expected struct {
		res      model.Order
		hasErr   bool
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx   = context.Background()
		uuid  = uuid.New()
		num   = gofakeit.Int64()
		items = []model.OrderItem{
			{Uuid: uuid, Type: model.PartTypeHull},
			{Uuid: uuid, Type: model.PartTypeEngine},
		}

		expModel = model.Order{
			OrderUUID:       uuid,
			Items:           items,
			TotalPrice:      num,
			TransactionUUID: nil,
			PaymentMethod:   nil,
			Status:          model.OrderStatusPendingPayment,
			CreatedAt:       time.Now(),
		}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.OrderRepository)
		expected  expected
	}{
		{
			name: "заказ найден - ошибок нет",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(mock.Anything, uuid).Return(expModel, nil)
			},
			expected: expected{
				res:    expModel,
				hasErr: false,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderNotFound",
			args: args{id: uuid},
			setupMock: func(r *mocks.OrderRepository) {
				r.EXPECT().Get(mock.Anything, uuid).Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expected: expected{
				res:    model.Order{},
				hasErr: true,
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
				r.EXPECT().Get(mock.Anything, uuid).Return(model.Order{}, errInternal)
			},
			expected: expected{
				res:    model.Order{},
				hasErr: true,
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

			test.setupMock(r)

			res, err := s.Get(ctx, test.args.id)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.Empty(t, res)
			} else {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.res.OrderUUID, res.OrderUUID)
				assert.Equal(t, len(test.expected.res.Items), len(res.Items))
				assert.Equal(t, test.expected.res.TotalPrice, res.TotalPrice)
				assert.Nil(t, res.TransactionUUID)
				assert.Nil(t, res.PaymentMethod)
				assert.Equal(t, test.expected.res.Status, res.Status)
				assert.Equal(t, test.expected.res.CreatedAt, res.CreatedAt)
			}
		})
	}
}
