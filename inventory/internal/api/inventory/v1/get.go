package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/inventory/internal/api/converter"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// GetPart возвращает деталь по UUID.
func (a *api) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	// Валидировать формат UUID → INVALID_ARGUMENT. Валидация формата UUID v4
	// Можно использовать github.com/google/uuid:
	id, err := uuid.Parse(req.GetUuid())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrPartInvalidUUID)
	}

	part, err := a.partService.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &inventoryv1.GetPartResponse{Part: converter.PartToProto(part)}, nil
}
