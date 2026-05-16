package input

import (
	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model/valueobject"
)

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []uuid.UUID
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType valueobject.PartType
}
