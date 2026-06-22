package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/auth"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Create(ctx context.Context, orderRq input.CreateOrderRq) (input.CreateOrderRs, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Create")
	defer span.End()

	for _, i := range orderRq.Items {
		span.SetAttributes(
			attribute.String("item.type", string(i.Type)),
			attribute.String("item.uuid", i.Uuid.String()),
		)
	}

	userUUID, ok := auth.UserUUIDFromContext(ctx)
	if !ok {
		return input.CreateOrderRs{}, sharedErrs.NewUnauthorized(errs.ErrUnauthorized)
	}

	rqParts := orderRq.PartUUIDsAsSlice()

	// 2. Получить детали через InventoryService.ListParts
	parts, err := s.inventoryClient.ListParts(ctx, rqParts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return input.CreateOrderRs{}, sharedErrs.NewNotFound(err) // клиент уже завернул в BusinessError у себя
	}

	// Запретите использование одного UUID в разных слотах → тоже ErrPartTypeMismatch.
	err = resolvedShipSlots(parts, orderRq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return input.CreateOrderRs{}, sharedErrs.NewInvalidArgument(err)
	}

	// Проверяет, что все детали существуют (иначе 404)
	if parts == nil || len(parts) < len(rqParts) {
		return input.CreateOrderRs{}, sharedErrs.NewNotFound(errs.ErrPartNotFound)
	}

	// 3. Проверить stock_quantity > 0
	// 4. Вычислить total_price
	var totalPrice int64
	for _, p := range parts {
		if p.StockQuantity <= 0 {
			err := fmt.Errorf("деталь %s: %w", *p.Name, errs.ErrOutOfStock)

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return input.CreateOrderRs{}, sharedErrs.NewConflict(err)
		}
		totalPrice += p.Price
	}

	// Запросить проверку на валидацию совместимости деталей между собой
	err = s.inventoryClient.ValidateCompatibility(ctx, parts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return input.CreateOrderRs{}, processGRPCError(err)
	}

	// Зарезервировать детали на складе
	err = s.inventoryClient.ReserveParts(ctx, rqParts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return input.CreateOrderRs{}, processGRPCError(err)
	}

	// 5. Сгенерировать order_uuid (UUID v4)
	// Создать заказ со статусом PENDING_PAYMENT
	order := model.NewOrderToCreate(parts, totalPrice, userUUID)

	// 6. Сохранить в store
	err = s.orderRepository.Create(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return input.CreateOrderRs{}, sharedErrs.NewInternal(err)
	}

	// Фиксируем метрику для отображения в Grafana
	ordersCreated.Add(ctx, 1)

	span.SetAttributes(
		attribute.String("order.uuid", order.OrderUUID.String()),
		attribute.Int("order.items_count", len(order.Items)),
		attribute.Int64("order.total_price", order.TotalPrice),
	)

	// 7. Вернуть order_uuid и total_price
	return input.CreateOrderRs{OrderUUID: order.OrderUUID, TotalPrice: totalPrice}, nil
}

// resolvedShipSlots проверяет каждую деталь, полученную из БД, на соответствие типа
// детали согласно запрошенным слотам.
func resolvedShipSlots(parts []model.OrderItem, orderRq input.CreateOrderRq) error {
	pMap := make(map[uuid.UUID]model.OrderItem, len(parts))
	for _, p := range parts {
		pMap[p.Uuid] = p
	}

	for _, i := range orderRq.Items {
		if p, ok := pMap[i.Uuid]; ok && p.Type != i.Type {
			return errs.ErrPartTypeMismatch
		}
	}

	return nil
}

func processGRPCError(err error) error {
	switch {
	case errors.Is(err, errs.ErrGrpcNotFound):
		return sharedErrs.NewNotFound(err)
	case errors.Is(err, errs.ErrGrpcInvalidArgument):
		return sharedErrs.NewInvalidArgument(err)
	case errors.Is(err, errs.ErrGrpcCancelledOperation):
		return sharedErrs.NewFailedPrecondition(err)
	default:
		return sharedErrs.NewInternal(err)
	}
}
