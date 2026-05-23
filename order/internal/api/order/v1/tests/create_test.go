package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/order/internal/api/converter"
	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	api "github.com/MikeKirillov/rocketservices/order/internal/api/order/v1"
	"github.com/MikeKirillov/rocketservices/order/internal/api/order/v1/mocks"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		req *orderv1.CreateOrderRequest
	}
	type expected struct {
		res      orderv1.CreateOrderRes
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		id      = uuid.New()
		orderId = uuid.New()
		num     = gofakeit.Int64()

		req = &orderv1.CreateOrderRequest{
			HullUUID:   id,
			EngineUUID: id,
		}
		reqOpt = &orderv1.CreateOrderRequest{
			HullUUID:   id,
			EngineUUID: id,
			ShieldUUID: orderv1.NewOptNilUUID(id),
			WeaponUUID: orderv1.NewOptNilUUID(id),
		}

		orderRq, _ = converter.OrderCreateRqToModel(req)
		orderRs    = input.CreateOrderRs{
			OrderUUID:  orderId,
			TotalPrice: num * 2,
		}
		orderRqOpt, _ = converter.OrderCreateRqToModel(reqOpt)
		orderRsOpt    = input.CreateOrderRs{
			OrderUUID:  orderId,
			TotalPrice: num * 4,
		}
		orderRsEmpty = input.CreateOrderRs{}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.OrderService)
		expected  expected
	}{
		{
			name: "успешное создание hull+engine - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Create(ctx, orderRq).Return(orderRs, nil)
			},
			expected: expected{
				res: &orderv1.CreateOrderResponse{
					OrderUUID:  orderRs.OrderUUID,
					TotalPrice: orderRs.TotalPrice,
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "успешное создание shield+weapon - ошибок нет",
			args: args{req: reqOpt},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Create(ctx, orderRqOpt).Return(orderRsOpt, nil)
			},
			expected: expected{
				res: &orderv1.CreateOrderResponse{
					OrderUUID:  orderRsOpt.OrderUUID,
					TotalPrice: orderRsOpt.TotalPrice,
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "одна из деталей не найдена - errs.ErrPartNotFound",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Create(ctx, orderRq).
					Return(orderRsEmpty, sharedErrs.NewNotFound(errs.ErrPartNotFound))
			},
			expected: expected{
				res: &orderv1.CreateOrderNotFound{
					Code:    http.StatusNotFound,
					Message: "деталь не найдена",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err) // тк в хендлере вернется не ошибка, а тело CreateOrderNotFound,
					// которая по сути является http-оберткой, убеждаемся, что не было прочих ошибок
				},
			},
		},
		{
			name: "деталь закончилась на складе	- errs.ErrOutOfStock",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Create(ctx, orderRq).
					Return(orderRsEmpty, sharedErrs.NewConflict(errs.ErrOutOfStock))
			},
			expected: expected{
				res: &orderv1.CreateOrderConflict{
					Code:    http.StatusConflict,
					Message: "деталь отсутствует на складе",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "ошибка при сохранении - errInternal",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Create(ctx, orderRq).
					Return(orderRsEmpty, sharedErrs.NewInternal(errInternal))
			},
			expected: expected{
				res: &orderv1.CreateOrderInternalServerError{
					Code:    http.StatusInternalServerError,
					Message: "внутренняя ошибка",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s := mocks.NewOrderService(t)
			h := api.NewOrderHandler(s)

			test.setupMock(s)

			res, err := h.CreateOrder(ctx, test.args.req)

			test.expected.errCheck(t, err)

			require.NotEmpty(t, res)
			assert.Equal(t, test.expected.res, res)
		})
	}
}
