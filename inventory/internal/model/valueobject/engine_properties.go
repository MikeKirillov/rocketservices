package valueobject

import (
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

type EngineClass string

const (
	EngineClassA EngineClass = "A"
	EngineClassB EngineClass = "B"
	EngineClassC EngineClass = "C"
)

func NewEngineClass(s string) (EngineClass, error) {
	t := EngineClass(s)
	switch t {
	case EngineClassA, EngineClassB, EngineClassC:
		return t, nil
	default:
		return "", errs.ErrInvalidProperties
	}
}

// EngineProperties — экспортируемый алиас для приватной структуры.
type EngineProperties = engineProperties

// engineProperties - свойства двигателя.
type engineProperties struct {
	class            EngineClass
	requiredStrength int
}

func (p *engineProperties) Class() EngineClass    { return p.class }
func (p *engineProperties) RequiredStrength() int { return p.requiredStrength }

func NewEngineProperties(class string, strength int) (PartProperties, error) {
	c, err := NewEngineClass(class)
	if err != nil {
		return PartProperties{}, fmt.Errorf("неизвестный класс двигателя, получено %s: %w", class, err)
	}
	if strength < 30 || strength > 200 {
		return PartProperties{}, fmt.Errorf(
			"прочность корпуса должна быть от 30 до 200, получено %d: %w",
			strength,
			errs.ErrInvalidProperties,
		)
	}
	return PartProperties{
		engine: &engineProperties{class: c, requiredStrength: strength},
	}, nil
}
