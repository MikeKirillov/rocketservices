package converter

import (
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

func PartToModel(part record.Part) model.Part {
	return model.Part{
		UUID:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      partTypeToModel(part.PartType),
		StockQuantity: part.StockQuantity,
		CreatedAt:     part.CreatedAt,
	}
}

func partTypeToModel(partType string) model.PartType {
	switch partType {
	case "HULL":
		return model.PartTypeHull

	case "ENGINE":
		return model.PartTypeEngine

	case "SHIELD":
		return model.PartTypeShield

	case "WEAPON":
		return model.PartTypeWeapon

	default:
		return model.PartTypeUnspecified
	}
}
