package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	api "github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1"
	"github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func TestCommitParts(t *testing.T) {
	t.Parallel()

	type args struct {
		req *inventoryv1.CommitPartsRequest
	}

	type expected struct {
		res      *inventoryv1.CommitPartsResponse
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()
		id  = gofakeit.UUID()
		uid = uuid.MustParse(id)

		req = &inventoryv1.CommitPartsRequest{Uuids: []string{id}}
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.PartService)
		expected  expected
	}{
		{
			name: "зарезервировано - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().Commit(ctx, []uuid.UUID{uid}).Return(nil)
			},
			expected: expected{
				res: &inventoryv1.CommitPartsResponse{},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "переданы невалидные id - errs.ErrPartInvalidUUID",
			args: args{req: &inventoryv1.CommitPartsRequest{Uuids: []string{"id"}}},
			setupMock: func(s *mocks.PartService) {
				// мок не требуется, тк ошибка вылезет раньше вызова интеграции
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errs.ErrPartInvalidUUID))
				},
			},
		},
		{
			name: "сервис валидации выбрасывает любую ошибку, на примере - errs.ErrOutOfStock",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Commit(ctx, []uuid.UUID{uid}).
					Return(sharedErrs.NewResourceExhausted(errs.ErrOutOfStock))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.ResourceExhausted, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errs.ErrOutOfStock))
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

			res, err := a.CommitParts(ctx, test.args.req)

			if test.expected.res == nil {
				test.expected.errCheck(t, err)
			} else {
				require.Empty(t, res)
			}
		})
	}
}
