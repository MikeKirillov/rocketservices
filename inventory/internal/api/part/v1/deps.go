package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/input"
)

type PartService interface {
	Get(ctx context.Context, uuid uuid.UUID) (entity.Part, error)
	List(ctx context.Context, filter input.PartFilter) ([]entity.Part, error)
}
