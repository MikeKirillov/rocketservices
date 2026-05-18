package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	api "github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1"
	"github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1/mocks"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/input"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func TestValidateCompatibility(t *testing.T) {
	t.Parallel()

	type args struct {
		req *inventoryv1.ValidateCompatibilityRequest
	}

	type expected struct {
		res      *inventoryv1.ValidateCompatibilityResponse
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()
		id  = gofakeit.UUID()
		uid = uuid.MustParse(id)

		req = &inventoryv1.ValidateCompatibilityRequest{
			HullUuid:   id,
			EngineUuid: id,
			ShieldUuid: "",
		}
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(s *mocks.PartService)
		expected  expected
	}{
		{
			name: "запрос на валидацию - ошибок нет",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Validate(ctx, mock.MatchedBy(func(s input.ShipSlots) bool {
						return s.HullUUID == uid && s.EngineUUID == uid &&
							s.ShieldUUID == uuid.Nil && s.WeaponUUID == uuid.Nil
					})).
					Return(nil)
			},
			expected: expected{
				res: &inventoryv1.ValidateCompatibilityResponse{},
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "ошибка валидации входящих uuid - errs.ErrPartInvalidUUID",
			args: args{req: &inventoryv1.ValidateCompatibilityRequest{
				HullUuid:   "",
				EngineUuid: "id",
				ShieldUuid: "",
			}},
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
			name: "сервис валидации выбрасывает любую ошибку, на примере - errs.ErrPartTypeMismatch",
			args: args{req: req},
			setupMock: func(s *mocks.PartService) {
				s.EXPECT().
					Validate(ctx, mock.MatchedBy(func(s input.ShipSlots) bool {
						return s.HullUUID == uid && s.EngineUUID == uid &&
							s.ShieldUUID == uuid.Nil && s.WeaponUUID == uuid.Nil
					})).
					Return(sharedErrs.NewInvalidArgument(errs.ErrPartTypeMismatch))
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)

					// Доменная ошибка - маппим Code -> grpc code, message - клиенту.
					be, ok := sharedErrs.AsBusinessError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.InvalidArgument, be.Code().GRPCCode())
					assert.True(t, errors.Is(err, errs.ErrPartTypeMismatch))
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

			res, err := a.ValidateCompatibility(ctx, test.args.req)

			if test.expected.res == nil {
				test.expected.errCheck(t, err)
			} else {
				require.Empty(t, res)
			}
		})
	}
}
