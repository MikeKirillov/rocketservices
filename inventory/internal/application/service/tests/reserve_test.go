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

func TestReserve(t *testing.T) {
	type args struct {
		ids []uuid.UUID
	}

	type expected struct {
		err      error
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx     = context.Background()
		id      = uuid.New()
		str     = gofakeit.Word()
		num64   = gofakeit.Int64()
		num     = gofakeit.Int()
		created = time.Now()
		hullId  = uuid.New()

		ids         = []uuid.UUID{hullId}
		errInternal = errors.New("внутренняя ошибка")

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

		expModelOutStock = model.RestorePart(
			id,
			str,
			str,
			num64,
			model.PartTypeHull,
			0,
			0,
			hullProperty,
			created,
		)
		expModelSliceOutStock = []model.Part{expModelOutStock}
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(m *mocks.TxManager, r *mocks.PartRepository)
		expected  expected
	}{
		{
			name: "зарезервировано - ошибок нет",
			args: args{ids: ids},
			setupMock: func(m *mocks.TxManager, r *mocks.PartRepository) {
				r.EXPECT().ListByUUIDs(ctx, ids).Return(expModelSlice, nil)
				r.EXPECT().UpdateReservedBatch(ctx, expModelSlice).Return(nil)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(nil)
			},
			expected: expected{
				err: nil,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "один из UUID не найден - errs.ErrPartNotFound",
			args: args{ids: ids},
			setupMock: func(m *mocks.TxManager, r *mocks.PartRepository) {
				r.EXPECT().ListByUUIDs(ctx, ids).Return(nil, errs.ErrPartNotFound)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errs.ErrPartNotFound)
			},
			expected: expected{
				err: errs.ErrPartNotFound,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartNotFound))
				},
			},
		},
		{
			name: "ошибка репозитория - errInternal",
			args: args{ids: ids},
			setupMock: func(m *mocks.TxManager, r *mocks.PartRepository) {
				r.EXPECT().ListByUUIDs(ctx, ids).Return(expModelSlice, nil)
				r.EXPECT().UpdateReservedBatch(ctx, expModelSlice).Return(errInternal)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errInternal)
			},
			expected: expected{
				err: errInternal,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
		{
			name: "ошибка при создании резерва - errs.ErrOutOfStock",
			args: args{ids: ids},
			setupMock: func(m *mocks.TxManager, r *mocks.PartRepository) {
				r.EXPECT().ListByUUIDs(ctx, ids).Return(expModelSliceOutStock, nil)
				m.EXPECT().
					Do(ctx, mock.Anything).
					Run(func(ctx context.Context, fn func(context.Context) error) {
						_ = fn(ctx) // запускает выполнение функции, содержащей бизнес-шаги
					}).
					Return(errs.ErrOutOfStock)
			},
			expected: expected{
				err: errs.ErrOutOfStock,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOutOfStock))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := mocks.NewPartRepository(t)
			c := mocks.NewCompatibilityChecker(t)
			m := mocks.NewTxManager(t)
			s := part.NewService(r, c, m)

			test.setupMock(m, r)

			err := s.Release(ctx, test.args.ids)

			test.expected.errCheck(t, err)
		})
	}
}
