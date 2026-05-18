package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/application/part"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/application/part/mocks"
)

func TestReserve(t *testing.T) {
	t.Parallel()

	type args struct {
		ids []uuid.UUID
	}

	type expected struct {
		err      error
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx = context.Background()

		hullId   = uuid.New()
		engineId = uuid.New()
		shieldId = uuid.New()
		weaponId = uuid.New()

		ids         = []uuid.UUID{hullId, engineId, shieldId, weaponId}
		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(m *mocks.TxManager)
		expected  expected
	}{
		{
			name: "зарезервировано - ошибок нет",
			args: args{ids: ids},
			setupMock: func(m *mocks.TxManager) {
				m.EXPECT().Do(ctx, mock.Anything).Return(nil)
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
			setupMock: func(m *mocks.TxManager) {
				m.EXPECT().Do(ctx, mock.Anything).Return(errs.ErrPartNotFound)
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
			setupMock: func(m *mocks.TxManager) {
				m.EXPECT().Do(ctx, mock.Anything).Return(errInternal)
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
			setupMock: func(m *mocks.TxManager) {
				m.EXPECT().Do(ctx, mock.Anything).Return(errs.ErrOutOfStock)
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
			t.Parallel()

			r := mocks.NewPartRepository(t)
			c := mocks.NewCompatibilityChecker(t)
			m := mocks.NewTxManager(t)
			s := part.NewService(r, c, m)

			test.setupMock(m)

			err := s.Release(ctx, test.args.ids)

			test.expected.errCheck(t, err)
		})
	}
}
