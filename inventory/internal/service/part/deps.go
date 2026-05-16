package part

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/valueobject"
)

type PartRepository interface {
	Get(ctx context.Context, uuid uuid.UUID) (entity.Part, error)
	List(ctx context.Context) ([]entity.Part, error)
	ListByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]entity.Part, error)
	ListByPartType(ctx context.Context, partType valueobject.PartType) ([]entity.Part, error)
}
