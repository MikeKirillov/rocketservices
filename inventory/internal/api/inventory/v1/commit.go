package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/api/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) CommitParts(ctx context.Context, req *inventoryv1.CommitPartsRequest) (*inventoryv1.CommitPartsResponse, error) {
	input, err := converter.IdsToUuid(req.GetUuids())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(err)
	}

	if err := a.partService.Commit(ctx, input); err != nil {
		return nil, err
	}

	return &inventoryv1.CommitPartsResponse{}, nil
}
