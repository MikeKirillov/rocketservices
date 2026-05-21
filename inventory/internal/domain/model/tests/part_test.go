package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

func TestReserve(t *testing.T) {
	var (
		id      = uuid.New()
		str     = gofakeit.Word()
		num64   = gofakeit.Int64()
		num     = gofakeit.Int()
		now     = time.Now()
		hull, _ = model.NewHullProperties(num)

		part    = model.RestorePart(id, str, str, num64, model.PartTypeHull, 100, 0, hull, now)
		partErr = model.RestorePart(id, str, str, num64, model.PartTypeHull, 0, 0, hull, now)
	)

	t.Run("отмена релиза - ошибок нет", func(t *testing.T) {
		err := part.Reserve()

		require.NoError(t, err)
	})

	t.Run("отмена релиза - errs.ErrOutOfStock", func(t *testing.T) {
		err := partErr.Reserve()

		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrOutOfStock))
	})
}

func TestRelease(t *testing.T) {
	var (
		id      = uuid.New()
		str     = gofakeit.Word()
		num64   = gofakeit.Int64()
		num     = gofakeit.Int()
		now     = time.Now()
		hull, _ = model.NewHullProperties(num)

		part    = model.RestorePart(id, str, str, num64, model.PartTypeHull, num, num, hull, now)
		partErr = model.RestorePart(id, str, str, num64, model.PartTypeHull, num, 0, hull, now)
	)

	t.Run("отмена релиза - ошибок нет", func(t *testing.T) {
		err := part.Release()

		require.NoError(t, err)
	})

	t.Run("отмена релиза - errs.ErrNothingToRelease", func(t *testing.T) {
		err := partErr.Release()

		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrNothingToRelease))
	})
}
