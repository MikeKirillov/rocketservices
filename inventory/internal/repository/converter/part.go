package converter

import (
	"encoding/json"
	"fmt"

	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/valueobject"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

func PartRecordToModel(rec record.Part) (entity.Part, error) {
	var propsRec record.PartPropertiesRecord
	if err := json.Unmarshal(rec.Properties, &propsRec); err != nil {
		return entity.Part{}, fmt.Errorf("десериализовать свойства: %w", err)
	}

	props, err := partPropertiesFromRecord(propsRec)
	if err != nil {
		return entity.Part{}, fmt.Errorf("конвертировать свойства: %w", err)
	}

	partType, err := valueobject.NewPartType(rec.PartType)
	if err != nil {
		return entity.Part{}, fmt.Errorf("конвертировать тип детали: %w", err)
	}

	return entity.RestorePart(
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

func partPropertiesFromRecord(rec record.PartPropertiesRecord) (valueobject.PartProperties, error) {
	switch {
	case rec.Hull != nil:
		return valueobject.NewHullProperties(rec.Hull.Strength)

	case rec.Engine != nil:
		return valueobject.NewEngineProperties(rec.Engine.Class, rec.Engine.RequiredStrength)

	case rec.Shield != nil:
		return valueobject.NewShieldProperties(rec.Shield.ShieldType)

	case rec.Weapon != nil:
		return valueobject.NewWeaponProperties(rec.Weapon.WeaponType)

	default:
		return valueobject.PartProperties{}, nil
	}
}

func PartsToModel(parts []record.Part) ([]entity.Part, error) {
	r := make([]entity.Part, 0, len(parts))
	for _, p := range parts {
		part, err := PartRecordToModel(p)
		if err != nil {
			return nil, err
		}
		r = append(r, part)
	}
	return r, nil
}
