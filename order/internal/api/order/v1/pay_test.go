package v1

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/order/internal/api/order/v1/mocks"
	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func TestPayOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		req    *orderv1.PayOrderRequest
		params orderv1.PayOrderParams
	}

	type expected struct {
		res      orderv1.PayOrderRes
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx    = context.Background()
		id     = uuid.New()
		tranId = uuid.New()

		req    = &orderv1.PayOrderRequest{PaymentMethod: orderv1.PaymentMethodCREDITCARD}
		params = orderv1.PayOrderParams{OrderUUID: id}

		payOrderRq, _ = converter.PayOrderRqToModel(params.OrderUUID, req.PaymentMethod)
		errInternal   = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.OrderService)
		expected  expected
	}{
		{
			name: "успешная оплата - ошибок нет",
			args: args{req: req, params: params},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Pay(ctx, payOrderRq).Return(tranId, nil)
			},
			expected: expected{
				res: &orderv1.PayOrderResponse{TransactionUUID: tranId},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderNotFound",
			args: args{req: req, params: params},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, payOrderRq).
					Return(uuid.Nil, sharedErrs.NewNotFound(errs.ErrOrderNotFound))
			},
			expected: expected{
				res: &orderv1.PayOrderNotFound{
					Code:    http.StatusNotFound,
					Message: "заказ не найден",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err) // тк в хендлере вернется не ошибка, а тело PayOrderNotFound,
					// которая по сути является http-оберткой, убеждаемся, что не было прочих ошибок
				},
			},
		},
		{
			name: "заказ уже оплачен или отменен - errs.ErrOrderAlreadyPaidOrCancalled",
			args: args{req: req, params: params},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, payOrderRq).
					Return(uuid.Nil, sharedErrs.NewConflict(errs.ErrOrderAlreadyPaidOrCancalled))
			},
			expected: expected{
				res: &orderv1.PayOrderConflict{
					Code:    http.StatusConflict,
					Message: "заказ уже оплачен или был отменен",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "ошибка PaymentService - errInternal",
			args: args{req: req, params: params},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, payOrderRq).
					Return(uuid.Nil, sharedErrs.NewInternal(errInternal))
			},
			expected: expected{
				res: &orderv1.PayOrderInternalServerError{
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
			h := NewOrderHandler(s)

			test.setupMock(s)

			res, err := h.PayOrder(ctx, test.args.req, test.args.params)

			test.expected.errCheck(t, err)

			require.NotEmpty(t, res)
			assert.Equal(t, test.expected.res, res)
		})
	}
}
