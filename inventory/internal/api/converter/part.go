package converter

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MikeKirillov/rocketservices/inventory/internal/application/input"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func PartToProto(part model.Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID().String(),
		Name:          part.Name(),
		Description:   part.Description(),
		Price:         part.Price(),
		PartType:      PartTypeToProto(part.PartType()),
		StockQuantity: int64(part.StockQuantity()),
		CreatedAt:     timestamppb.New(part.CreatedAt()),
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

func PartsRqToModel(req *inventoryv1.ListPartsRequest) (input.PartFilter, error) {
	uuids, err := IdsToUuid(req.GetUuids())
	if err != nil {
		return input.PartFilter{}, err
	}

	return input.PartFilter{
		UUIDs:    uuids,
		PartType: partTypeToModel(req.GetPartType()),
	}, nil
}

func IdsToUuid(ids []string) ([]uuid.UUID, error) {
	if ids == nil {
		return nil, nil
	}

	uuids := make([]uuid.UUID, 0, len(ids))

	for _, v := range ids {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("деталь %s: %w", v, errs.ErrPartInvalidUUID)
		}

		uuids = append(uuids, id)
	}
	return uuids, nil
}

func partTypeToModel(partType inventoryv1.PartType) model.PartType {
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
		return ""
	}
}

func ValidateCompToModel(req *inventoryv1.ValidateCompatibilityRequest) (input.ShipSlots, error) {
	hullUuid, err := parseRequired(req.GetHullUuid())
	if err != nil {
		return input.ShipSlots{}, fmt.Errorf("идентификатор корпуса %s: %w", req.GetHullUuid(), err)
	}

	engineUuid, err := parseRequired(req.GetEngineUuid())
	if err != nil {
		return input.ShipSlots{}, fmt.Errorf("идентификатор двигателя %s: %w", req.GetEngineUuid(), err)
	}

	shieldUuid, err := parseOptional(req.GetShieldUuid())
	if err != nil {
		return input.ShipSlots{}, fmt.Errorf("идентификатор защиты %s: %w", req.GetShieldUuid(), err)
	}

	weaponUuid, err := parseOptional(req.GetWeaponUuid())
	if err != nil {
		return input.ShipSlots{}, fmt.Errorf("идентификатор вооружения %s: %w", req.GetWeaponUuid(), err)
	}

	return input.ShipSlots{
		HullUUID:   hullUuid,
		EngineUUID: engineUuid,
		ShieldUUID: shieldUuid,
		WeaponUUID: weaponUuid,
	}, nil
}

func parseRequired(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errs.ErrPartInvalidUUID
	}
	return id, nil
}

func parseOptional(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return parseRequired(s)
}
