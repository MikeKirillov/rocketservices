package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/inventory/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func (a *api) ListParts(ctx context.Context, req *inventoryv1.ListPartsRequest) (*inventoryv1.ListPartsResponse, error) {
	parts, err := a.partService.List(ctx, converter.PartsRqToModel(req))
	if err != nil {
		if errors.Is(err, errs.ErrPartInvalidUUID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "внутрення ошибка: %v", err)
	}

	partsRs := make([]*inventoryv1.Part, 0, len(parts))
	for _, part := range parts {
		partsRs = append(partsRs, converter.PartToProto(part))
	}
	return &inventoryv1.ListPartsResponse{Parts: partsRs}, nil
}
