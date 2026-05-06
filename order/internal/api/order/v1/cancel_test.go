package v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/order/internal/api/order/v1/mocks"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func TestCancelOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		req orderv1.CancelOrderParams
	}

	type expected struct {
		res      orderv1.CancelOrderRes
		hasErr   bool
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()
		id  = uuid.New()

		req = orderv1.CancelOrderParams{OrderUUID: id}
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.OrderService)
		expected  expected
	}{
		{
			name: "успешная отмена - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Cancel(ctx, id).Return(nil)
			},
			expected: expected{
				res:    &orderv1.CancelOrderResponse{},
				hasErr: false,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ не найден - errs.ErrOrderNotFound",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Cancel(ctx, id).
					Return(sharedErrs.NewNotFound(errs.ErrOrderNotFound))
			},
			expected: expected{
				res: &orderv1.CancelOrderNotFound{
					Code:    http.StatusNotFound,
					Message: "заказ не найден",
				},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err) // тк в хендлере вернется не ошибка, а тело CancelOrderNotFound,
					// которая по сути является http-оберткой, убеждаемся, что не было прочих ошибок
				},
			},
		},
		{
			name: "заказ уже отменен - errs.ErrOrderAlreadyCancelled",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Cancel(ctx, id).
					Return(sharedErrs.NewConflict(errs.ErrOrderAlreadyCancelled))
			},
			expected: expected{
				res: &orderv1.CancelOrderConflict{
					Code:    http.StatusConflict,
					Message: "заказ уже отменён",
				},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "заказ уже оплачен - errs.ErrOrderAlreadyPaid",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Cancel(ctx, id).
					Return(sharedErrs.NewConflict(errs.ErrOrderAlreadyPaid))
			},
			expected: expected{
				res: &orderv1.CancelOrderConflict{
					Code:    http.StatusConflict,
					Message: "заказ уже оплачен",
				},
				hasErr: true,
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

			res, err := h.CancelOrder(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.NotEmpty(t, res)
			} else {
				require.Empty(t, res)
			}
			assert.Equal(t, test.expected.res, res)
		})
	}
}
