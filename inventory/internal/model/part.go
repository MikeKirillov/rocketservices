package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

// Part представляет деталь космического корабля.
type Part struct {
	UUID          uuid.UUID
	Name          string
	Description   string
	Price         int64 // в копейках
	PartType      PartType
	StockQuantity int64
	CreatedAt     time.Time
}

type PartType string

const (
	PartTypeHull   PartType = "HULL"
	PartTypeEngine PartType = "ENGINE"
	PartTypeShield PartType = "SHIELD"
	PartTypeWeapon PartType = "WEAPON"
)

func NewPartType(s string) (PartType, error) {
	pt := PartType(s)
	switch pt {
	case PartTypeHull, PartTypeEngine, PartTypeShield, PartTypeWeapon:
		return pt, nil
	default:
		return "", fmt.Errorf("деталь %s: %w", s, errs.ErrWrongPartType)
	}
}

func OrderPartsByIds(parts []Part, ids []uuid.UUID) []Part {
	pMap := make(map[uuid.UUID]Part)
	for _, p := range parts {
		pMap[p.UUID] = p
	}

	pOrdered := make([]Part, 0, len(ids))
	for _, id := range ids {
		if p, ok := pMap[id]; ok {
			pOrdered = append(pOrdered, p)
		}
	}

	return pOrdered
}

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []uuid.UUID
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType PartType
}
