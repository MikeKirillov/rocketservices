package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

func TestNewWeaponType(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.WeaponType
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "laser",
			expected: expected{
				result: model.WeaponTypeLaser,
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

			res, err := model.NewWeaponType(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res)
			}
		})
	}
}

func TestNewWeaponProperties(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.WeaponType
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "laser",
			expected: expected{
				result: model.WeaponTypeLaser,
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

			res, err := model.NewWeaponProperties(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res.Weapon().WeaponType())
			}
		})
	}
}
