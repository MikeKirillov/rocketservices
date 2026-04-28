package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func PartToProto(part model.Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      PartTypeToProto(part.PartType),
		StockQuantity: part.StockQuantity,
		CreatedAt:     timestamppb.New(part.CreatedAt),
	}
}

func PartsRqToModel(req *inventoryv1.ListPartsRequest) model.PartFilter {
	return model.PartFilter{
		UUIDs:    req.Uuids,
		PartType: PartTypeToModel(req.PartType),
	}
}

func PartTypeToProto(partType model.PartType) inventoryv1.PartType {
	switch partType {
	case model.PartTypeHull:
		return inventoryv1.PartType_PART_TYPE_HULL

	case model.PartTypeEngine:
		return inventoryv1.PartType_PART_TYPE_ENGINE

	case model.PartTypeShield:
		return inventoryv1.PartType_PART_TYPE_SHIELD

	case model.PartTypeWeapon:
		return inventoryv1.PartType_PART_TYPE_WEAPON

	default:
		return inventoryv1.PartType_PART_TYPE_UNSPECIFIED
	}
}

func PartTypeToModel(partType inventoryv1.PartType) model.PartType {
	switch partType {
	case inventoryv1.PartType_PART_TYPE_HULL:
		return model.PartTypeHull

	case inventoryv1.PartType_PART_TYPE_ENGINE:
		return model.PartTypeEngine

	case inventoryv1.PartType_PART_TYPE_SHIELD:
		return model.PartTypeShield

	case inventoryv1.PartType_PART_TYPE_WEAPON:
		return model.PartTypeWeapon

	default:
		return model.PartTypeUnspecified
	}
}
