package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/MikeKirillov/rocketservices/inventory/internal/api/converter"
	api "github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1"
	"github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func TestGetPart(t *testing.T) {
	t.Parallel()

	type args struct {
		req *inventoryv1.GetPartRequest
	}

	type expected struct {
		res      *inventoryv1.GetPartResponse
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx   = context.Background()
		id    = uuid.New()
		str   = gofakeit.Word()
		num64 = gofakeit.Int64()
		num   = gofakeit.Int()

		created = time.Now()

		req = &inventoryv1.GetPartRequest{
			Uuid: id.String(),
		}

		hullProperty, _ = model.NewHullProperties(num)
		expModel        = model.RestorePart(
			id,
			str,
			str,
			num64,
			model.PartTypeHull,
			num,
			num,
			hullProperty,
			created,
		)

		expProto    = converter.PartToProto(expModel)
		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.PartService)
		expected  expected
	}{
		{
			name: "деталь найдена - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().Get(ctx, id).Return(expModel, nil)
			},
			expected: expected{
				res: &inventoryv1.GetPartResponse{Part: expProto},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "деталь не найдена - codes.NotFound",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Get(ctx, id).
					Return(model.Part{}, sharedErrs.NewNotFound(errs.ErrPartNotFound))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.NotFound, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errs.ErrPartNotFound))
				},
			},
		},
		{
			name: "пустой UUID - codes.InvalidArgument",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Get(ctx, id).
					Return(model.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartOmittedUUID))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errs.ErrPartOmittedUUID))
				},
			},
		},
		{
			name: "ошибка репозитория - codes.Internal",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Get(ctx, id).
					Return(model.Part{}, sharedErrs.NewInternal(errInternal))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.Internal, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s := mocks.NewPartService(t)
			a := api.NewAPI(s)

			test.setupMock(s)

			res, err := a.GetPart(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.res == nil {
				require.Nil(t, res)
			} else {
				require.NotNil(t, res)
				assert.Equal(t, test.expected.res.GetPart().GetUuid(), res.GetPart().GetUuid())
				assert.Equal(t, test.expected.res.GetPart().GetName(), res.GetPart().GetName())
				assert.Equal(t, test.expected.res.GetPart().GetDescription(), res.GetPart().GetDescription())
				assert.Equal(t, test.expected.res.GetPart().GetPrice(), res.GetPart().GetPrice())
				assert.Equal(t, test.expected.res.GetPart().GetPartType(), res.GetPart().GetPartType())
				assert.Equal(t, test.expected.res.GetPart().GetStockQuantity(), res.GetPart().GetStockQuantity())
				assert.Equal(t, test.expected.res.GetPart().GetCreatedAt(), res.GetPart().GetCreatedAt())
			}
		})
	}
}
