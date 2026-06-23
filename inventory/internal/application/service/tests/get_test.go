package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	part "github.com/MikeKirillov/rocketservices/inventory/internal/application/service"
	"github.com/MikeKirillov/rocketservices/inventory/internal/application/service/mocks"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id uuid.UUID
	}

	type expected struct {
		res      model.Part
		hasErr   bool
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		uuid    = uuid.New()
		str     = gofakeit.Word()
		num64   = gofakeit.Int64()
		num     = gofakeit.Int()
		created = time.Now()

		hullProperty, _ = model.NewHullProperties(num)
		expModel        = model.RestorePart(
			uuid,
			str,
			str,
			num64,
			model.PartTypeHull,
			num,
			num,
			hullProperty,
			created,
		)

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.PartRepository)
		expected  expected
	}{
		{
			name: "деталь найдена - ошибок нет",
			args: args{id: uuid},
			setupMock: func(r *mocks.PartRepository) {
				r.EXPECT().Get(mock.Anything, uuid).Return(expModel, nil)
			},
			expected: expected{
				res:    expModel,
				hasErr: false,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "деталь не найдена - errs.ErrPartNotFound",
			args: args{id: uuid},
			setupMock: func(r *mocks.PartRepository) {
				r.EXPECT().Get(mock.Anything, uuid).Return(model.Part{}, errs.ErrPartNotFound)
			},
			expected: expected{
				res:    model.Part{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartNotFound))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{id: uuid},
			setupMock: func(r *mocks.PartRepository) {
				r.EXPECT().Get(mock.Anything, uuid).Return(model.Part{}, errInternal)
			},
			expected: expected{
				res:    model.Part{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := mocks.NewPartRepository(t)
			c := mocks.NewCompatibilityChecker(t)
			m := mocks.NewTxManager(t)
			s := part.NewService(r, c, m)

			test.setupMock(r)

			res, err := s.Get(ctx, test.args.id)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.Empty(t, res)
			} else {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.res.UUID(), res.UUID())
				assert.Equal(t, test.expected.res.Name(), res.Name())
				assert.Equal(t, test.expected.res.Description(), res.Description())
				assert.Equal(t, test.expected.res.Price(), res.Price())
				assert.Equal(t, test.expected.res.PartType(), res.PartType())
				assert.Equal(t, test.expected.res.StockQuantity(), res.StockQuantity())
				assert.Equal(t, test.expected.res.CreatedAt(), res.CreatedAt())
			}
		})
	}
}
