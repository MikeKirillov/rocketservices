package part

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/part/mocks"
)

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		filter model.PartFilter
	}

	type expected struct {
		res      []model.Part
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		id      = uuid.New()
		id2     = uuid.New()
		str     = gofakeit.Word()
		num     = gofakeit.Int64()
		created = time.Now()

		filter = model.PartFilter{
			UUIDs: nil,
		}
		filterHull = model.PartFilter{
			PartType: model.PartTypeHull,
			UUIDs:    nil,
		}
		filterUuid = model.PartFilter{
			UUIDs: []uuid.UUID{id},
		}
		filterUuids = model.PartFilter{
			UUIDs: []uuid.UUID{id, id2},
		}

		expModel = []model.Part{
			{
				UUID:          id,
				Name:          "дорогая деталь",
				Description:   str,
				Price:         num,
				PartType:      model.PartTypeHull,
				StockQuantity: num,
				CreatedAt:     created,
			},
		}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.PartRepository)
		expected  expected
	}{
		{
			name: "без фильтра (все детали) - ошибок нет",
			args: args{filter: filter},
			setupMock: func(r *mocks.PartRepository) {
				r.EXPECT().List(ctx).Return(expModel, nil)
			},
			expected: expected{
				res: expModel,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "фильтр по типу HULL - все детали с типом HULL",
			args: args{filter: filterHull},
			setupMock: func(s *mocks.PartRepository) {
				s.EXPECT().ListByPartType(ctx, model.PartTypeHull).Return(expModel, nil)
			},
			expected: expected{
				res: expModel,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "фильтр по конкретным UUID - детали в порядке запрошенных",
			args: args{filter: filterUuid},
			setupMock: func(s *mocks.PartRepository) {
				s.EXPECT().ListByUUIDs(ctx, []uuid.UUID{id}).Return(expModel, nil)
			},
			expected: expected{
				res: expModel,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "один из UUID не найден - errs.ErrPartNotFound",
			args: args{filter: filterUuids},
			setupMock: func(s *mocks.PartRepository) {
				s.EXPECT().
					ListByUUIDs(ctx, []uuid.UUID{id, id2}).
					Return(nil, errs.ErrPartNotFound)
			},
			expected: expected{
				res: nil,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartNotFound))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{filter: filterHull},
			setupMock: func(s *mocks.PartRepository) {
				s.EXPECT().
					ListByPartType(ctx, model.PartTypeHull).
					Return(nil, errInternal)
			},
			expected: expected{
				res: nil,
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
			s := NewService(r)

			test.setupMock(r)

			res, err := s.List(ctx, test.args.filter)

			test.expected.errCheck(t, err)

			if test.expected.res == nil {
				require.Nil(t, res)
			} else {
				require.NotNil(t, res)
				assert.Equal(t, 1, len(res))
				assert.Equal(t, test.expected.res[0].Name, res[0].Name)
				assert.Equal(t, test.expected.res[0].PartType, res[0].PartType)
				assert.Equal(t, test.expected.res[0].UUID, res[0].UUID)
			}
		})
	}
}
