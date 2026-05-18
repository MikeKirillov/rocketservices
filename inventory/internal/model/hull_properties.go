package model

import (
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

// HullProperties — экспортируемый алиас для приватной структуры.
type HullProperties = hullProperties

// hullProperties - свойства корпуса.
type hullProperties struct {
	strength int
}

// NewHullProperties создаёт свойства корпуса. Прочность должна быть в диапазоне 30–200.
func NewHullProperties(strength int) (PartProperties, error) {
	if strength < 30 || strength > 200 {
		return PartProperties{}, fmt.Errorf(
			"прочность корпуса должна быть от 30 до 200, получено %d: %w",
			strength,
			errs.ErrInvalidProperties,
		)
	}
	return PartProperties{
		hull: &hullProperties{strength: strength},
	}, nil
}

func (p *hullProperties) Strength() int { return p.strength }

// CanSupport проверяет, выдержит ли корпус нагрузку двигателя.
func (p *hullProperties) CanSupport(e *engineProperties) bool {
	return p.strength >= e.requiredStrength
}
