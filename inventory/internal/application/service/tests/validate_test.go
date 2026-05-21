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

	"github.com/MikeKirillov/rocketservices/inventory/internal/application/input"
	part "github.com/MikeKirillov/rocketservices/inventory/internal/application/service"
	"github.com/MikeKirillov/rocketservices/inventory/internal/application/service/mocks"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

func TestValidate(t *testing.T) {
	type args struct {
		slots input.ShipSlots
	}

	type expected struct {
		err      error
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx      = context.Background()
		hullId   = uuid.New()
		engineId = uuid.New()
		shieldId = uuid.New()
		weaponId = uuid.New()
		str      = gofakeit.Word()
		num64    = gofakeit.Int64()
		num      = gofakeit.Int()
		now      = time.Now()

		slots    = input.ShipSlots{HullUUID: hullId, EngineUUID: engineId, ShieldUUID: shieldId, WeaponUUID: weaponId}
		slotsRep = input.ShipSlots{HullUUID: hullId, EngineUUID: hullId, ShieldUUID: shieldId, WeaponUUID: weaponId}

		hp, _ = model.NewHullProperties(num)
		hrp   = model.RestorePart(hullId, str, str, num64, model.PartTypeHull, num, num, hp, now)

		ep, _ = model.NewEngineProperties(string(model.EngineClassA), num)
		erp   = model.RestorePart(engineId, str, str, num64, model.PartTypeEngine, num, num, ep, now)

		sp, _ = model.NewShieldProperties(string(model.ShieldTypeEnergy))
		srp   = model.RestorePart(shieldId, str, str, num64, model.PartTypeShield, num, num, sp, now)

		wp, _ = model.NewWeaponProperties(string(model.WeaponTypeLaser))
		wrp   = model.RestorePart(weaponId, str, str, num64, model.PartTypeWeapon, num, num, wp, now)

		wp_miss, _ = model.NewWeaponProperties(string(model.WeaponTypeLaser))
		wrpp_miss  = model.RestorePart(weaponId, str, str, num64, model.PartTypeHull, num, num, wp_miss, now)

		parts     = []model.Part{hrp, erp, srp, wrp}
		partsMiss = []model.Part{hrp, erp, srp, wrpp_miss}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.PartRepository, c *mocks.CompatibilityChecker)
		expected  expected
	}{
		{
			name: "повторяющиеся идентификаторы hull-engine - errs.ErrPartTypeMismatch",
			args: args{slots: slotsRep},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				// ошибка выпадает до вызова интеграций, потому моки не требуются
			},
			expected: expected{
				err: errs.ErrPartTypeMismatch,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartTypeMismatch))
				},
			},
		},
		{
			name: "получены все слоты - ошибок нет",
			args: args{slots: slots},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				r.EXPECT().
					ListByUUIDs(ctx, mock.MatchedBy(func(ids []uuid.UUID) bool { // тк последовательность id при
						// вызове метода ListByUUIDs() может быть изменена после работы с мапой, используем mock.MatchedBy(),
						// в которй мы провряем каждый id, что он равен хотя бы одном из id запроса
						for _, id := range ids {
							return id == slots.HullUUID || id == slots.EngineUUID ||
								id == slots.ShieldUUID || id == slots.WeaponUUID
						}
						return true
					})).
					Return(parts, nil)
				c.EXPECT().
					Check(mock.MatchedBy(func(s model.ResolvedShipSlots) bool {
						return s.Hull.UUID() == slots.HullUUID && s.Hull.PartType() == model.PartTypeHull &&
							s.Engine.UUID() == slots.EngineUUID && s.Engine.PartType() == model.PartTypeEngine &&
							s.Shield.UUID() == slots.ShieldUUID && s.Shield.PartType() == model.PartTypeShield &&
							s.Weapon.UUID() == slots.WeaponUUID && s.Weapon.PartType() == model.PartTypeWeapon
					})).
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
			args: args{slots: slots},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				r.EXPECT().
					ListByUUIDs(ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
						for _, id := range ids {
							return id == slots.HullUUID || id == slots.EngineUUID ||
								id == slots.ShieldUUID || id == slots.WeaponUUID
						}
						return true
					})).
					Return(nil, errs.ErrPartNotFound)
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
			args: args{slots: slots},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				r.EXPECT().
					ListByUUIDs(ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
						for _, id := range ids {
							return id == slots.HullUUID || id == slots.EngineUUID ||
								id == slots.ShieldUUID || id == slots.WeaponUUID
						}
						return true
					})).
					Return(nil, errInternal)
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
			name: "нарушение правил совместимости - errs.ErrIncompatibleParts",
			args: args{slots: slots},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				r.EXPECT().
					ListByUUIDs(ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
						for _, id := range ids {
							return id == slots.HullUUID || id == slots.EngineUUID ||
								id == slots.ShieldUUID || id == slots.WeaponUUID
						}
						return true
					})).
					Return(parts, nil)
				c.EXPECT().
					Check(mock.MatchedBy(func(s model.ResolvedShipSlots) bool {
						return s.Hull.UUID() == slots.HullUUID && s.Hull.PartType() == model.PartTypeHull &&
							s.Engine.UUID() == slots.EngineUUID && s.Engine.PartType() == model.PartTypeEngine &&
							s.Shield.UUID() == slots.ShieldUUID && s.Shield.PartType() == model.PartTypeShield &&
							s.Weapon.UUID() == slots.WeaponUUID && s.Weapon.PartType() == model.PartTypeWeapon
					})).
					Return(errs.ErrIncompatibleParts)
			},
			expected: expected{
				err: errs.ErrIncompatibleParts,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrIncompatibleParts))
				},
			},
		},
		{
			name: "несовместимые детали из БД согласно запрошенным слотам - errs.ErrPartTypeMismatch",
			args: args{slots: slots},
			setupMock: func(r *mocks.PartRepository, c *mocks.CompatibilityChecker) {
				r.EXPECT().
					ListByUUIDs(ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
						for _, id := range ids {
							return id == slots.HullUUID || id == slots.EngineUUID ||
								id == slots.ShieldUUID || id == slots.WeaponUUID
						}
						return true
					})).
					Return(partsMiss, nil)
			},
			expected: expected{
				err: errs.ErrPartTypeMismatch,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartTypeMismatch))
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

			test.setupMock(r, c)

			err := s.Validate(ctx, test.args.slots)

			test.expected.errCheck(t, err)
		})
	}
}
