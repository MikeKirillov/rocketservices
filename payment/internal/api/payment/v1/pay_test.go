package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/payment/internal/api/payment/v1/mocks"
	"github.com/MikeKirillov/rocketservices/payment/internal/converter"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func TestPayOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		req *paymentv1.PayOrderRequest
	}

	type expected struct {
		res      *paymentv1.PayOrderResponse
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()
		id  = uuid.New()

		reqCard = &paymentv1.PayOrderRequest{
			OrderUuid:     id.String(),
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
		}

		reqSbp = &paymentv1.PayOrderRequest{
			OrderUuid:     id.String(),
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_SBP,
		}

		modelCard, _ = converter.PaymentToModel(reqCard)
		modelSbp, _  = converter.PaymentToModel(reqSbp)
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.OrderService)
		expected  expected
	}{
		{
			name: "успешная оплата заказа картой",
			args: args{req: reqCard},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Pay(ctx, modelCard).Return(id, nil)
			},
			expected: expected{
				res: &paymentv1.PayOrderResponse{TransactionUuid: id.String()},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "успешная оплата заказа СБП",
			args: args{req: reqSbp},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Pay(ctx, modelSbp).Return(id, nil)
			},
			expected: expected{
				res: &paymentv1.PayOrderResponse{TransactionUuid: id.String()},
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
			a := NewAPI(s)

			test.setupMock(s)

			res, err := a.PayOrder(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.res != nil {
				require.NotNil(t, res)
				assert.NotNil(t, res.GetTransactionUuid())
				assert.Equal(t, test.expected.res.GetTransactionUuid(), res.GetTransactionUuid())
			} else {
				require.Nil(t, res)
			}
		})
	}
}
