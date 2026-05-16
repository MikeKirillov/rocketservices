package converter

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/valueobject"
	"github.com/MikeKirillov/rocketservices/inventory/internal/service/input"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func PartToProto(part entity.Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID().String(),
		Name:          part.Name(),
		Description:   part.Description(),
		Price:         part.Price(),
		PartType:      PartTypeToProto(part.PartType()),
		StockQuantity: part.StockQuantity(),
		CreatedAt:     timestamppb.New(part.CreatedAt()),
	}
}

func PartTypeToProto(partType valueobject.PartType) inventoryv1.PartType {
	switch partType {
	case valueobject.PartTypeHull:
		return inventoryv1.PartType_PART_TYPE_HULL

	case valueobject.PartTypeEngine:
		return inventoryv1.PartType_PART_TYPE_ENGINE

	case valueobject.PartTypeShield:
		return inventoryv1.PartType_PART_TYPE_SHIELD

	case valueobject.PartTypeWeapon:
		return inventoryv1.PartType_PART_TYPE_WEAPON

	default:
		return inventoryv1.PartType_PART_TYPE_UNSPECIFIED
	}
}

func PartsRqToModel(req *inventoryv1.ListPartsRequest) (input.PartFilter, error) {
	var uuids []uuid.UUID

	if req.Uuids != nil {
		uuids = make([]uuid.UUID, 0, len(req.Uuids))
		for _, v := range req.Uuids {
			id, err := uuid.Parse(v)
			if err != nil {
				return input.PartFilter{}, fmt.Errorf("деталь %s: %w", v, errs.ErrPartInvalidUUID)
			}
			uuids = append(uuids, id)
		}
	}
	return input.PartFilter{
		UUIDs:    uuids,
		PartType: partTypeToModel(req.PartType),
	}, nil
}

func partTypeToModel(partType inventoryv1.PartType) valueobject.PartType {
	switch partType {
	case inventoryv1.PartType_PART_TYPE_HULL:
		return valueobject.PartTypeHull

	case inventoryv1.PartType_PART_TYPE_ENGINE:
		return valueobject.PartTypeEngine

	case inventoryv1.PartType_PART_TYPE_SHIELD:
		return valueobject.PartTypeShield

	case inventoryv1.PartType_PART_TYPE_WEAPON:
		return valueobject.PartTypeWeapon

	default:
		return ""
	}
}
