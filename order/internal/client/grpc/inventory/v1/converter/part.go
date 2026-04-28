package converter

import (
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func PartToModel(part *inventoryv1.Part) model.Part {
	return model.Part{
		Uuid:          part.GetUuid(),
		Name:          part.GetName(),
		Price:         part.GetPrice(),
		StockQuantity: part.GetStockQuantity(),
	}
}
