package converter

import (
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

func PartsToModel(parts []record.Part) []model.Part {
	r := make([]model.Part, 0, len(parts))
	for _, p := range parts {
		r = append(r, partToModel(p))
	}
	return r
}

func partToModel(p record.Part) model.Part {
	return model.Part{
		UUID:          p.UUID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		PartType:      model.PartType(p.PartType),
		StockQuantity: p.StockQuantity,
		CreatedAt:     p.CreatedAt,
	}
}
