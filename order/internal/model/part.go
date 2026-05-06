package model

import (
	"github.com/google/uuid"
)

type Part struct {
	Uuid          uuid.UUID
	Type          PartType
	Name          *string
	Price         int64
	StockQuantity int64
}

type PartType string

const (
	PartTypeHull   PartType = "HULL"
	PartTypeEngine PartType = "ENGINE"
	PartTypeShield PartType = "SHIELD"
	PartTypeWeapon PartType = "WEAPON"
)
