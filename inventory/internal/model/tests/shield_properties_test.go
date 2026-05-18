package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

func TestNewShieldType(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.ShieldType
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "energy",
			expected: expected{
				result: model.ShieldTypeEnergy,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name:  "конвертация с ошибкой - errs.ErrInvalidProperties",
			input: "yo",
			expected: expected{
				result: "",
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

			res, err := model.NewShieldType(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res)
			}
		})
	}
}

func TestNewShieldProperties(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.ShieldType
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "energy",
			expected: expected{
				result: model.ShieldTypeEnergy,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name:  "конвертация с ошибкой - errs.ErrInvalidProperties",
			input: "yo",
			expected: expected{
				result: "",
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

			res, err := model.NewShieldProperties(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res.Shield().ShieldType())
			}
		})
	}
}

func TestConflictsWith(t *testing.T) {
	t.Run("совместимость щита и вооружения - котфликта нет", func(t *testing.T) {
		s, err := model.NewShieldProperties("energy")
		w, errw := model.NewWeaponProperties("laser")

		require.NoError(t, err)
		require.NoError(t, errw)

		r := s.Shield().ConflictsWith(w.Weapon())
		assert.False(t, r)
	})

	t.Run("совместимость щита и вооружения - котфликт", func(t *testing.T) {
		s, err := model.NewShieldProperties("plasma")
		w, errw := model.NewWeaponProperties("laser")

		require.NoError(t, err)
		require.NoError(t, errw)

		r := s.Shield().ConflictsWith(w.Weapon())
		assert.True(t, r)
	})
}
