package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/converter"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) ListParts(ctx context.Context, req *inventoryv1.ListPartsRequest) (*inventoryv1.ListPartsResponse, error) {
	parts, err := a.partService.List(ctx, converter.PartsRqToModel(req))
	if err != nil {
		return nil, err
	}

	partsRs := make([]*inventoryv1.Part, 0, len(parts))
	for _, part := range parts {
		partsRs = append(partsRs, converter.PartToProto(part))
	}
	return &inventoryv1.ListPartsResponse{Parts: partsRs}, nil
}
