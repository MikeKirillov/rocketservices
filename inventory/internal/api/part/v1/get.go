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

// GetPart возвращает деталь по UUID.
func (a *api) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	part, err := a.partService.Get(ctx, req.GetUuid())
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPartOmittedUUID) || errors.Is(err, errs.ErrPartInvalidUUID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, errs.ErrPartNotFound):
			return nil, status.Errorf(codes.NotFound, "деталь %s не найдена", req.GetUuid())
		default:
			return nil, status.Errorf(codes.Internal, "внутрення ошибка: %v", err)
		}
	}
	return &inventoryv1.GetPartResponse{Part: converter.PartToProto(part)}, nil
}
