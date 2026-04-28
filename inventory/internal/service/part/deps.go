package part

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

type PartRepository interface {
	Get(ctx context.Context, uuid string) (model.Part, error)
	List(ctx context.Context) ([]model.Part, error)
	ListByUUIDs(ctx context.Context, uuids []string) ([]model.Part, error)
	ListByPartType(ctx context.Context, partType model.PartType) ([]model.Part, error)
}
