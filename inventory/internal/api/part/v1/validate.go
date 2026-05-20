package v1

import (
	"context"

	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) ValidateCompatibility(ctx context.Context, req *inventoryv1.ValidateCompatibilityRequest) (*inventoryv1.ValidateCompatibilityResponse, error) {
	// TODO
	return nil, nil
}
