package input

import (
	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
)

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []uuid.UUID
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType model.PartType
}

type ShipSlots struct {
	HullUUID   uuid.UUID
	EngineUUID uuid.UUID
	ShieldUUID uuid.UUID // optional
	WeaponUUID uuid.UUID // optional
}
