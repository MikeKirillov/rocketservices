package part

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
)

type CompatibilityChecker interface {
	Check(slots model.ResolvedShipSlots) error
}

type PartRepository interface {
	Get(ctx context.Context, uuid uuid.UUID) (model.Part, error)
	List(ctx context.Context) ([]model.Part, error)
	ListByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error)
	ListByUUIDsForUpdate(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error)
	ListByPartType(ctx context.Context, partType model.PartType) ([]model.Part, error)
	UpdateReservedBatch(ctx context.Context, parts []model.Part) error
	UpdateCommitBatch(ctx context.Context, parts []model.Part) error
}
