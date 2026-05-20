package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
)

type PartService interface {
	Get(ctx context.Context, uuid uuid.UUID) (model.Part, error)
	List(ctx context.Context, filter model.PartFilter) ([]model.Part, error)
}
