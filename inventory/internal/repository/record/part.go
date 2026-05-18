package record

import (
	"time"

	"github.com/google/uuid"
)

// PartRecord — плоская структура для маппинга строки из БД.
type Part struct {
	UUID          uuid.UUID `db:"uuid"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	Price         int64     `db:"price"`
	PartType      string    `db:"part_type"`
	StockQuantity int       `db:"stock_quantity"`
	Reserved      int       `db:"reserved"`
	Properties    []byte    `db:"properties"` // JSONB из PostgreSQL
	CreatedAt     time.Time `db:"created_at"`
}

// PartPropertiesRecord — структура для десериализации JSONB.
// Живёт в repository-слое: json-теги не принадлежат доменной модели.
type PartPropertiesRecord struct {
	Hull   *HullPropertiesRecord   `json:"hull,omitempty"`
	Engine *EnginePropertiesRecord `json:"engine,omitempty"`
	Shield *ShieldPropertiesRecord `json:"shield,omitempty"`
	Weapon *WeaponPropertiesRecord `json:"weapon,omitempty"`
}

type HullPropertiesRecord struct {
	Strength int `json:"strength"`
}

type EnginePropertiesRecord struct {
	Class            string `json:"class"`
	RequiredStrength int    `json:"required_strength"`
}

type ShieldPropertiesRecord struct {
	ShieldType string `json:"shield_type"`
}

type WeaponPropertiesRecord struct {
	WeaponType string `json:"weapon_type"`
}
