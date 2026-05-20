package converter

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func PartToModel(part *inventoryv1.Part) (model.OrderItem, error) {
	id, err := uuid.Parse(part.GetUuid())
	if err != nil {
		return model.OrderItem{}, errs.ErrInvalidUUID
	}
	pt, err := partTypeToModel(part.PartType)
	if err != nil {
		return model.OrderItem{}, err
	}
	return model.OrderItem{
		Uuid:          id,
		Type:          pt,
		Name:          new(part.GetName()),
		Price:         part.GetPrice(),
		StockQuantity: part.GetStockQuantity(),
	}, nil
}

func partTypeToModel(partType inventoryv1.PartType) (model.PartType, error) {
	switch partType {
	case inventoryv1.PartType_PART_TYPE_HULL:
		return model.PartTypeHull, nil

	case inventoryv1.PartType_PART_TYPE_ENGINE:
		return model.PartTypeEngine, nil

	case inventoryv1.PartType_PART_TYPE_SHIELD:
		return model.PartTypeShield, nil

	case inventoryv1.PartType_PART_TYPE_WEAPON:
		return model.PartTypeWeapon, nil

	default:
		return "", errs.ErrWrongPartType
	}
}

func PartUuidsToStr(partsUuids []uuid.UUID) []string {
	ids := make([]string, 0, len(partsUuids))
	for _, id := range partsUuids {
		ids = append(ids, id.String())
	}
	return ids
}

func OrderToPartUuids(order model.Order) []string {
	ids := make([]string, 0, len(order.Items))
	for _, i := range order.Items {
		ids = append(ids, i.Uuid.String())
	}
	return ids
}

// GrpcErrToDomain маппит gRPC ошибок в доменные ошибки.
func GrpcErrToDomain(st *status.Status, err error) error {
	switch st.Code() {
	case codes.NotFound:
		return errs.ErrGrpcNotFound

	case codes.FailedPrecondition, codes.ResourceExhausted:
		return fmt.Errorf("%w: %w", errs.ErrGrpcCancelledOperation, err)

	case codes.InvalidArgument:
		return errs.ErrGrpcInvalidArgument

	case codes.DeadlineExceeded: // Таймаут — клиент может повторить запрос
		slog.Error("таймаут", "message", st.Message())
		return errs.ErrGrpcTimeout

	case codes.Unavailable: // gRPC сервис недоступен
		slog.Error("gRPC сервис недоступен", "message", st.Message())
		return errs.ErrGrpcUnavailable

	case codes.Canceled: // Запрос был отменён (например, клиент закрыл соединение)
		slog.Error("запрос отменен", "message", st.Message())
		return errs.ErrGrpcUnavailable

	default:
		slog.Error("ошибка gRPC", "code", st.Code(), "message", st.Message())
		return errs.ErrGrpcInternal
	}
}

func ItemsToProto(items []model.OrderItem) *inventoryv1.ValidateCompatibilityRequest {
	r := &inventoryv1.ValidateCompatibilityRequest{}

	for _, i := range items {
		id := i.Uuid.String()
		switch i.Type {
		case model.PartTypeHull:
			r.HullUuid = id
		case model.PartTypeEngine:
			r.EngineUuid = id
		case model.PartTypeShield:
			r.ShieldUuid = id
		case model.PartTypeWeapon:
			r.WeaponUuid = id
		}
	}

	return r
}
