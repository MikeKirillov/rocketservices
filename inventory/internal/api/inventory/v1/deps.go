package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/input"
)

type PartService interface {
	Get(ctx context.Context, id uuid.UUID) (model.Part, error)
	List(ctx context.Context, filter input.PartFilter) ([]model.Part, error)
	Validate(ctx context.Context, slots input.ShipSlots) error
	Reserve(ctx context.Context, ids []uuid.UUID) error
	Release(ctx context.Context, ids []uuid.UUID) error
}
