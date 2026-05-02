package part

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/part/mocks"
)

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	type expected struct {
		res      model.Part
		hasErr   bool
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		uuid    = gofakeit.UUID()
		str     = gofakeit.Word()
		num     = gofakeit.Int64()
		created = time.Now()

		expModel = model.Part{
			UUID:          uuid,
			Name:          str,
			Description:   str,
			Price:         num,
			PartType:      model.PartTypeHull,
			StockQuantity: num,
			CreatedAt:     created,
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
			name: "деталь найдена - ошибок нет",
			args: args{id: uuid},
			setupMock: func(r *mocks.PartRepository) {
				r.EXPECT().Get(ctx, uuid).Return(expModel, nil)
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
				r.EXPECT().Get(ctx, uuid).Return(model.Part{}, errs.ErrPartNotFound)
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
				r.EXPECT().Get(ctx, uuid).Return(model.Part{}, errInternal)
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
			s := NewService(r)

			test.setupMock(r)

			res, err := s.Get(ctx, test.args.id)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.Empty(t, res)
			} else {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.res.UUID, res.UUID)
				assert.Equal(t, test.expected.res.Name, res.Name)
				assert.Equal(t, test.expected.res.Description, res.Description)
				assert.Equal(t, test.expected.res.Price, res.Price)
				assert.Equal(t, test.expected.res.PartType, res.PartType)
				assert.Equal(t, test.expected.res.StockQuantity, res.StockQuantity)
				assert.Equal(t, test.expected.res.CreatedAt, res.CreatedAt)
			}
		})
	}
}
