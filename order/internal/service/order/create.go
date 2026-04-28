package order

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
)

type OrderStatus string

func (s *service) Create(ctx context.Context, orderRq model.CreateOrderRq) (model.CreateOrderRs, error) {
	// 1. Валидация: hull_uuid и engine_uuid обязательны
	if err := checkRequiredField(orderRq.HullUUID, "hull_uuid"); err != nil {
		return model.CreateOrderRs{}, fmt.Errorf("поле hull_uuid: %w", err)
	}
	if err := checkRequiredField(orderRq.EngineUUID, "engine_uuid"); err != nil {
		return model.CreateOrderRs{}, fmt.Errorf("поле engine_uuid: %w", err)
	}

	rqParts := orderRq.PartUUIDsAsStringSlice()
	err := validateUUIDs(rqParts)
	if err != nil {
		return model.CreateOrderRs{}, err
	}

	// 2. Получить детали через InventoryService.ListParts
	parts, err := s.inventoryClient.ListParts(ctx, rqParts)
	if err != nil {
		return model.CreateOrderRs{}, err
	}

	// Проверяет, что все детали существуют (иначе 404)
	if parts == nil || len(parts) < len(rqParts) {
		slog.Error("детали найдены", "запрашивали", strconv.Itoa(len(rqParts)), "найдено", strconv.Itoa(len(parts)))
		return model.CreateOrderRs{}, errs.ErrPartNotFound
	}

	// 3. Проверить stock_quantity > 0
	// 4. Вычислить total_price
	var totalPrice int64
	for _, p := range parts {
		if p.StockQuantity <= 0 {
			slog.Error("недостаточное количество на складе", "деталь", p.Name, "количество", p.StockQuantity)
			return model.CreateOrderRs{}, fmt.Errorf("деталь %s: %w", p.Name, errs.ErrOutOfStock)
		}
		totalPrice += p.Price
	}

	// 5. Сгенерировать order_uuid (UUID v4)
	orderUUID := uuid.New()

	// 6. Создать заказ со статусом PENDING_PAYMENT
	order := converter.OrderToRepoCreate(orderRq, totalPrice, orderUUID.String(), model.OrderStatusPendingPayment)

	// 7. Сохранить в store
	err = s.orderRepository.Create(ctx, order)
	if err != nil {
		return model.CreateOrderRs{}, fmt.Errorf("создать заявку: %w", err)
	}

	// 8. Вернуть order_uuid и total_price
	return model.CreateOrderRs{OrderUUID: orderUUID, TotalPrice: totalPrice}, nil
}

func checkRequiredField(field, fieldName string) error {
	if field == "" {
		slog.Error("некорректный запрос", "не передано обязательное поле", fieldName)
		return errs.ErrOmittedRequiredField
	}
	return nil
}

func validateUUIDs(parts []string) error {
	for _, id := range parts {
		if _, err := uuid.Parse(id); err != nil {
			slog.Error("неверный формат", "uuid", id)
			return fmt.Errorf("деталь %s: %w", id, errs.ErrInvalidUUID)
		}
	}
	return nil
}
