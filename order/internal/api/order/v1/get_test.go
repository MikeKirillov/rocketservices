package v1

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/order/internal/api/order/v1/mocks"
	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func TestGetOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		req orderv1.GetOrderParams
	}

	type expected struct {
		res      orderv1.GetOrderRes
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()
		id  = uuid.New()
		id1 = uuid.New()
		id2 = uuid.New()
		num = gofakeit.Int64()

		req = orderv1.GetOrderParams{OrderUUID: id}

		expModel = model.Order{
			OrderUUID: id1,
			Parts: []model.Part{
				{
					Uuid: id2,
					Type: model.PartTypeHull,
				},
				{
					Uuid: id2,
					Type: model.PartTypeEngine,
				},
			},
			TotalPrice: num,
			Status:     model.OrderStatusPendingPayment,
			CreatedAt:  time.Now(),
		}
		emptyModel = model.Order{}

		res, _      = converter.OrderToDTO(expModel)
		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.OrderService)
		expected  expected
	}{
		{
			name: "заказ найден - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Get(ctx, id).Return(expModel, nil)
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
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Get(ctx, id).
					Return(emptyModel, sharedErrs.NewNotFound(errs.ErrOrderNotFound))
			},
			expected: expected{
				res: &orderv1.GetOrderNotFound{
					Code:    http.StatusNotFound,
					Message: "заказ не найден",
				},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err) // тк в хендлере вернется не ошибка, а тело GetOrderNotFound,
					// которая по сути является http-оберткой, убеждаемся, что не было прочих ошибок
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{req: req},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Get(ctx, id).
					Return(emptyModel, sharedErrs.NewInternal(errInternal))
			},
			expected: expected{
				res: &orderv1.GetOrderInternalServerError{
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

			res, err := h.GetOrder(ctx, test.args.req)

			test.expected.errCheck(t, err)

			require.NotEmpty(t, res)
			assert.Equal(t, test.expected.res, res)
		})
	}
}
