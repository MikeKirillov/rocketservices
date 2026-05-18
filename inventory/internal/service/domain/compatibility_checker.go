package domain

import (
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

// componentSet хранит типизированные свойства комплектующих, извлечённые из набора.
type componentSet struct {
	hull   *model.HullProperties
	engine *model.EngineProperties
	shield *model.ShieldProperties
	weapon *model.WeaponProperties
}

// CompatibilityChecker проверяет совместимость деталей космического корабля.
type CompatibilityChecker struct{}

func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{}
}

// Check проверяет бизнес-правила совместимости для набора деталей.
func (c *CompatibilityChecker) Check(slots model.ResolvedShipSlots) error {
	// 1. Извлеките HullProperties и EngineProperties — они гарантированно есть, типы уже проверены.
	set := extractProperties(slots)

	// 2. Проверьте правило 2 через HullProperties.CanSupport — если нарушено, верните ErrIncompatibleParts.
	if !set.hull.CanSupport(set.engine) {
		return errs.ErrIncompatibleParts
	}

	// 3. Если оба опциональных слота заполнены — проверьте правило 3 через ShieldProperties.ConflictsWith.
	if set.shield != nil && set.weapon != nil && set.shield.ConflictsWith(set.weapon) {
		return errs.ErrIncompatibleParts
	}

	return nil
}

// extractProperties разбирает срез комплектующих и заполняет componentSet.
func extractProperties(slots model.ResolvedShipSlots) componentSet {
	var set componentSet

	hp := slots.Hull.Properties()
	if p := hp.Hull(); p != nil {
		set.hull = p
	}

	ep := slots.Engine.Properties()
	if p := ep.Engine(); p != nil {
		set.engine = p
	}

	if slots.Shield != nil {
		sp := slots.Shield.Properties()
		if p := sp.Shield(); p != nil {
			set.shield = p
		}
	}

	if slots.Weapon != nil {
		wp := slots.Weapon.Properties()
		if p := wp.Weapon(); p != nil {
			set.weapon = p
		}
	}

	return set
}
