package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/converter"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// GetPart возвращает деталь по UUID.
func (a *api) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	part, err := a.partService.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return &inventoryv1.GetPartResponse{Part: converter.PartToProto(part)}, nil
}
