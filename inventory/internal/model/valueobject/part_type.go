package valueobject

import (
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

type PartType string

const (
	PartTypeHull   PartType = "HULL"
	PartTypeEngine PartType = "ENGINE"
	PartTypeShield PartType = "SHIELD"
	PartTypeWeapon PartType = "WEAPON"
)

// NewPartType создаёт тип детали с валидацией.
func NewPartType(s string) (PartType, error) {
	pt := PartType(s)
	switch pt {
	case PartTypeHull, PartTypeEngine, PartTypeShield, PartTypeWeapon:
		return pt, nil
	default:
		return "", fmt.Errorf("деталь %s: %w", s, errs.ErrWrongPartType)
	}
}
