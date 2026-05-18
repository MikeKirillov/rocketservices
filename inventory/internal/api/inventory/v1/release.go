package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/api/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) ReleaseParts(ctx context.Context, req *inventoryv1.ReleasePartsRequest) (*inventoryv1.ReleasePartsResponse, error) {
	input, err := converter.IdsToUuid(req.GetUuids())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(err)
	}

	if err := a.partService.Release(ctx, input); err != nil {
		return nil, err
	}

	return &inventoryv1.ReleasePartsResponse{}, nil
}
