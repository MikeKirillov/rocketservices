package converter

import (
	"encoding/json"
	"fmt"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

func PartRecordToModel(rec record.Part) (model.Part, error) {
	var propsRec record.PartPropertiesRecord
	if err := json.Unmarshal(rec.Properties, &propsRec); err != nil {
		return model.Part{}, fmt.Errorf("десериализовать свойства: %w", err)
	}

	props, err := partPropertiesFromRecord(propsRec)
	if err != nil {
		return model.Part{}, fmt.Errorf("конвертировать свойства: %w", err)
	}

	partType, err := model.NewPartType(rec.PartType)
	if err != nil {
		return model.Part{}, fmt.Errorf("конвертировать тип детали: %w", err)
	}

	return model.RestorePart(
		rec.UUID,
		rec.Name,
		rec.Description,
		rec.Price,
		partType,
		rec.StockQuantity,
		rec.Reserved,
		props,
		rec.CreatedAt,
	), nil
}

func partPropertiesFromRecord(rec record.PartPropertiesRecord) (model.PartProperties, error) {
	switch {
	case rec.Hull != nil:
		return model.NewHullProperties(rec.Hull.Strength)

	case rec.Engine != nil:
		return model.NewEngineProperties(rec.Engine.Class, rec.Engine.RequiredStrength)

	case rec.Shield != nil:
		return model.NewShieldProperties(rec.Shield.ShieldType)

	case rec.Weapon != nil:
		return model.NewWeaponProperties(rec.Weapon.WeaponType)

	default:
		return model.PartProperties{}, nil
	}
}

func PartsToModel(parts []record.Part) ([]model.Part, error) {
	r := make([]model.Part, 0, len(parts))
	for _, p := range parts {
		part, err := PartRecordToModel(p)
		if err != nil {
			return nil, err
		}
		r = append(r, part)
	}
	return r, nil
}
