package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

func TestNewHullProperties(t *testing.T) {
	t.Parallel()

	var (
		engine, _     = model.NewEngineProperties("A", 100)
		engineMore, _ = model.NewEngineProperties("A", 200)
	)

	type args struct {
		strength int
		engine   *model.EngineProperties
	}

	type expected struct {
		strength   int
		errCheck   func(t *testing.T, err error)
		canSupport func(t *testing.T, h *model.HullProperties, e *model.EngineProperties)
	}

	tests := []struct {
		name     string
		input    args
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок + проверка прочности корпуса = true",
			input: args{strength: 100, engine: engine.Engine()},
			expected: expected{
				strength: 100,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
				canSupport: func(t *testing.T, h *model.HullProperties, e *model.EngineProperties) {
					require.True(t, h.CanSupport(e))
				},
			},
		},
		{
			name:  "конвертация типа без ошибок + проверка прочности корпуса = true",
			input: args{strength: 100, engine: engineMore.Engine()},
			expected: expected{
				strength: 100,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
				canSupport: func(t *testing.T, h *model.HullProperties, e *model.EngineProperties) {
					require.False(t, h.CanSupport(e))
				},
			},
		},
		{
			name:  "проверка прочности корпуса с ошибкой - errs.ErrInvalidProperties",
			input: args{strength: 10, engine: engine.Engine()},
			expected: expected{
				strength: 0,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrInvalidProperties))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res, err := model.NewHullProperties(test.input.strength)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.strength, res.Hull().Strength())

				test.expected.canSupport(t, res.Hull(), test.input.engine)
			} else {
				require.Empty(t, res)
			}
		})
	}
}
