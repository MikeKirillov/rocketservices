package converter

import (
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

func PartToModel(part record.Part) (model.Part, error) {
	pt, err := model.NewPartType(part.PartType)
	if err != nil {
		return model.Part{}, err
	}
	return model.Part{
		UUID:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      pt,
		StockQuantity: part.StockQuantity,
		CreatedAt:     part.CreatedAt,
	}, nil
}
