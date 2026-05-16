package valueobject

import (
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

type WeaponType string

const (
	WeaponTypeLaser   WeaponType = "laser"
	WeaponTypeMissile WeaponType = "missile"
)

func NewWeaponType(s string) (WeaponType, error) {
	t := WeaponType(s)
	switch t {
	case WeaponTypeLaser, WeaponTypeMissile:
		return t, nil
	default:
		return "", errs.ErrInvalidProperties
	}
}

// WeaponProperties — экспортируемый алиас для приватной структуры.
type WeaponProperties = weaponProperties

// weaponProperties - свойства оружия.
type weaponProperties struct {
	weaponType WeaponType
}

func (p *weaponProperties) WeaponType() WeaponType { return p.weaponType }

func NewWeaponProperties(t string) (PartProperties, error) {
	weaponType, err := NewWeaponType(t)
	if err != nil {
		return PartProperties{}, fmt.Errorf("неизвестный тип оружия, получено %s: %w", t, err)
	}
	return PartProperties{
		weapon: &weaponProperties{weaponType: weaponType},
	}, nil
}
