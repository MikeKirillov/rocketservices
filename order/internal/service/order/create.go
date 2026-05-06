package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Create(ctx context.Context, orderRq model.CreateOrderRq) (model.CreateOrderRs, error) {
	rqParts := orderRq.PartUUIDsAsSlice()

	// 2. Получить детали через InventoryService.ListParts
	parts, err := s.inventoryClient.ListParts(ctx, rqParts)
	if err != nil {
		return model.CreateOrderRs{}, sharedErrs.NewNotFound(err) // клиент уже завернул в BusinessError у себя
	}

	// Проверяет, что все детали существуют (иначе 404)
	if parts == nil || len(parts) < len(rqParts) {
		return model.CreateOrderRs{}, sharedErrs.NewNotFound(errs.ErrPartNotFound)
	}

	// 3. Проверить stock_quantity > 0
	// 4. Вычислить total_price
	var totalPrice int64
	for _, p := range parts {
		if p.StockQuantity <= 0 {
			return model.CreateOrderRs{}, sharedErrs.NewConflict(fmt.Errorf("деталь %s: %w", *p.Name, errs.ErrOutOfStock))
		}
		totalPrice += p.Price
	}

	// 5. Сгенерировать order_uuid (UUID v4)
	// Создать заказ со статусом PENDING_PAYMENT
	order := converter.CreateOrderRqToOrder(orderRq, totalPrice, model.OrderStatusPendingPayment, uuid.New(), time.Now())

	// 6. Сохранить в store
	orderUUID, err := s.orderRepository.Create(ctx, order)
	if err != nil {
		return model.CreateOrderRs{}, sharedErrs.NewInternal(err)
	}

	// 7. Вернуть order_uuid и total_price
	return model.CreateOrderRs{OrderUUID: orderUUID, TotalPrice: totalPrice}, nil
}
