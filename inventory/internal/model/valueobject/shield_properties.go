package valueobject

import (
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

type ShieldType string

const (
	ShieldTypeEnergy ShieldType = "energy"
	ShieldTypePlasma ShieldType = "plasma"
)

func NewShieldType(s string) (ShieldType, error) {
	t := ShieldType(s)
	switch t {
	case ShieldTypeEnergy, ShieldTypePlasma:
		return t, nil
	default:
		return "", errs.ErrInvalidProperties
	}
}

// ShieldProperties — экспортируемый алиас для приватной структуры.
type ShieldProperties = shieldProperties

// shieldProperties - свойства щита.
type shieldProperties struct {
	shieldType ShieldType
}

func (p *shieldProperties) ShieldType() ShieldType { return p.shieldType }

func NewShieldProperties(t string) (PartProperties, error) {
	shildType, err := NewShieldType(t)
	if err != nil {
		return PartProperties{}, fmt.Errorf("неизвестный тип щита, получено %s: %w", t, err)
	}
	return PartProperties{
		shield: &shieldProperties{shieldType: shildType},
	}, nil
}
