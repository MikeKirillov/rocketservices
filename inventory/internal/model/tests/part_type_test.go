package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

func TestNewPartType(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.PartType
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "HULL",
			expected: expected{
				result: model.PartTypeHull,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name:  "конвертация с ошибкой - errs.ErrWrongPartType",
			input: "yo",
			expected: expected{
				result: "",
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrWrongPartType))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res, err := model.NewPartType(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res)
			}
		})
	}
}
