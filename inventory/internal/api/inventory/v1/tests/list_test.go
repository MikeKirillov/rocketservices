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
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func TestListParts(t *testing.T) {
	t.Parallel()

	type args struct {
		req *inventoryv1.ListPartsRequest
	}

	type expected struct {
		res      *inventoryv1.ListPartsResponse
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		id      = uuid.New()
		uuid    = gofakeit.UUID()
		uuid2   = gofakeit.UUID()
		str     = gofakeit.Word()
		num64   = gofakeit.Int64()
		num     = gofakeit.Int()
		created = time.Now()

		req = &inventoryv1.ListPartsRequest{
			PartType: inventoryv1.PartType_PART_TYPE_UNSPECIFIED,
			Uuids:    nil,
		}
		reqHull = &inventoryv1.ListPartsRequest{
			PartType: inventoryv1.PartType_PART_TYPE_HULL,
			Uuids:    nil,
		}
		reqUuid = &inventoryv1.ListPartsRequest{
			Uuids: []string{uuid},
		}
		reqUuids = &inventoryv1.ListPartsRequest{
			Uuids: []string{uuid, uuid2},
		}

		rqModel, _      = converter.PartsRqToModel(req)
		rqModelHull, _  = converter.PartsRqToModel(reqHull)
		rqModelUuid, _  = converter.PartsRqToModel(reqUuid)
		rqModelUuids, _ = converter.PartsRqToModel(reqUuids)

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
		expModelSlice = []model.Part{expModel}

		expProto    = []*inventoryv1.Part{converter.PartToProto(expModelSlice[0])}
		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.PartService)
		expected  expected
	}{
		{
			name: "без фильтра (все детали) - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().List(ctx, rqModel).Return(expModelSlice, nil)
			},
			expected: expected{
				res: &inventoryv1.ListPartsResponse{Parts: expProto},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "фильтр по типу HULL - все детали с типом HULL",
			args: args{req: reqHull},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().List(ctx, rqModelHull).Return(expModelSlice, nil)
			},
			expected: expected{
				res: &inventoryv1.ListPartsResponse{Parts: expProto},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "фильтр по конкретным UUID - детали в порядке запрошенных",
			args: args{req: reqUuid},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().List(ctx, rqModelUuid).Return(expModelSlice, nil)
			},
			expected: expected{
				res: &inventoryv1.ListPartsResponse{Parts: expProto},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "один из UUID не найден - codes.NotFound",
			args: args{req: reqUuids},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					List(ctx, rqModelUuids).
					Return(nil, sharedErrs.NewNotFound(errs.ErrPartNotFound))
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
			name: "ошибка репозитория - codes.Internal",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					List(ctx, rqModel).
					Return(nil, sharedErrs.NewInternal(errInternal))
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

			res, err := a.ListParts(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.res == nil {
				require.Nil(t, res)
			} else {
				require.NotNil(t, res)
				assert.Equal(t, 1, len(res.GetParts()))
				assert.Equal(t, test.expected.res.GetParts()[0].GetName(), res.GetParts()[0].GetName())
				assert.Equal(t, test.expected.res.GetParts()[0].GetPartType(), res.GetParts()[0].GetPartType())
				assert.Equal(t, test.expected.res.GetParts()[0].GetUuid(), res.GetParts()[0].GetUuid())
			}
		})
	}
}
