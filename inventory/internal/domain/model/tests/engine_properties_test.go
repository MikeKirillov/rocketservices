package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

func TestNewEngineClass(t *testing.T) {
	t.Parallel()

	type expected struct {
		result   model.EngineClass
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: "A",
			expected: expected{
				result: model.EngineClassA,
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

			res, err := model.NewEngineClass(test.input)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res)
			}
		})
	}
}

func TestNewEngineProperties(t *testing.T) {
	t.Parallel()

	type args struct {
		class    string
		strength int
	}

	type expected struct {
		result   model.EngineClass
		strength int
		errCheck func(t *testing.T, err error)
	}

	tests := []struct {
		name     string
		input    args
		expected expected
	}{
		{
			name:  "конвертация типа без ошибок",
			input: args{class: "A", strength: 100},
			expected: expected{
				result:   model.EngineClassA,
				strength: 100,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name:  "конвертация типа с ошибкой - errs.ErrInvalidProperties",
			input: args{class: "yo", strength: 100},
			expected: expected{
				result:   "",
				strength: 0,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrInvalidProperties))
				},
			},
		},
		{
			name:  "проверка прочности корпуса с ошибкой - errs.ErrInvalidProperties",
			input: args{class: "yo", strength: 10},
			expected: expected{
				result:   "",
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

			res, err := model.NewEngineProperties(test.input.class, test.input.strength)

			test.expected.errCheck(t, err)

			if err == nil {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.result, res.Engine().Class())
				assert.Equal(t, test.expected.strength, res.Engine().RequiredStrength())
			} else {
				require.Empty(t, res)
			}
		})
	}
}
