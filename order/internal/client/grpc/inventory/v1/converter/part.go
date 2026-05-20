package converter

import (
	"github.com/google/uuid"

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
