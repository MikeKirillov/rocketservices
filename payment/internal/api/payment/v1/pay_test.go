package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/MikeKirillov/rocketservices/payment/internal/api/payment/v1/mocks"
	"github.com/MikeKirillov/rocketservices/payment/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/payment/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
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
		ctx  = context.Background()
		uuid = gofakeit.UUID()

		reqCard = &paymentv1.PayOrderRequest{
			OrderUuid:     uuid,
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
		}

		reqSbp = &paymentv1.PayOrderRequest{
			OrderUuid:     uuid,
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_SBP,
		}

		reqInvalidUuid = &paymentv1.PayOrderRequest{
			OrderUuid:     gofakeit.Word(),
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
		}

		reqEmptyUuid = &paymentv1.PayOrderRequest{
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
		}

		reqUnspecPmt = &paymentv1.PayOrderRequest{
			OrderUuid:     uuid,
			PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED,
		}

		modelCard        = converter.OrderToModel(reqCard)
		modelSbp         = converter.OrderToModel(reqSbp)
		modelInvalidUuid = converter.OrderToModel(reqInvalidUuid)
		modelEmptyUuid   = converter.OrderToModel(reqEmptyUuid)
		modelUnspecPmt   = converter.OrderToModel(reqUnspecPmt)
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
				s.EXPECT().Pay(ctx, modelCard).Return(uuid, nil)
			},
			expected: expected{
				res: &paymentv1.PayOrderResponse{TransactionUuid: uuid},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "успешная оплата заказа СБП",
			args: args{req: reqSbp},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().Pay(ctx, modelSbp).Return(uuid, nil)
			},
			expected: expected{
				res: &paymentv1.PayOrderResponse{TransactionUuid: uuid},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "невалидный order_uuid - codes.InvalidArgument",
			args: args{req: reqInvalidUuid},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, modelInvalidUuid).
					Return("", sharedErrs.NewInvalidArgument(errs.ErrOrderInvalidUUID))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					require.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(be.Unwrap(), errs.ErrOrderInvalidUUID))
				},
			},
		},
		{
			name: "пустой order_uuid - codes.InvalidArgument",
			args: args{req: reqEmptyUuid},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, modelEmptyUuid).
					Return("", sharedErrs.NewInvalidArgument(errs.ErrOrderOmittedUUID))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(be, errs.ErrOrderOmittedUUID))
				},
			},
		},
		{
			name: "PaymentMethod = UNSPECIFIED - codes.InvalidArgument",
			args: args{
				req: reqUnspecPmt,
			},
			setupMock: func(s *mocks.OrderService) {
				s.EXPECT().
					Pay(ctx, modelUnspecPmt).
					Return("", sharedErrs.NewInvalidArgument(errs.ErrOrderInvalidPayMethod))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(be, errs.ErrOrderInvalidPayMethod))
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
