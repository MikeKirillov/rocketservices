package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/inventory/internal/api/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) ValidateCompatibility(ctx context.Context, req *inventoryv1.ValidateCompatibilityRequest) (*inventoryv1.ValidateCompatibilityResponse, error) {
	shipSlots, err := converter.ValidateCompToModel(req)
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(err)
	}

	err = a.partService.Validate(ctx, shipSlots)
	if err != nil {
		return nil, err
	}
	return &inventoryv1.ValidateCompatibilityResponse{}, nil
}
