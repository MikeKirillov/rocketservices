package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/middleware"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

type OrderStatus string

const (
	pendingPayment OrderStatus = "PENDING_PAYMENT"
	paid           OrderStatus = "PAID"
	cancelled      OrderStatus = "CANCELLED"
)

const (
	// Таймаут для gRPC вызовов.
	grpcCallTimeout = 5 * time.Second
)

// Order представляет заказ на постройку космического корабля.
type Order struct {
	OrderUUID       uuid.UUID
	HullUUID        uuid.UUID
	EngineUUID      uuid.UUID
	ShieldUUID      *uuid.UUID // опциональный
	WeaponUUID      *uuid.UUID // опциональный
	TotalPrice      int64      // в копейках
	TransactionUUID *uuid.UUID
	PaymentMethod   *string
	Status          string // PENDING_PAYMENT, PAID, CANCELLED
	CreatedAt       time.Time
}

// OrderStore — хранилище заказов (in-memory).
type OrderStore struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]Order
}

// NewOrderStore создаёт новое пустое хранилище заказов.
func NewOrderStore() *OrderStore {
	return &OrderStore{
		orders: make(map[uuid.UUID]Order),
	}
}

func SetupRouter(h *OrderHandler, middlewareTimeout time.Duration) (http.Handler, error) {
	// Создать OpenAPI сервер
	orderServer, err := SetupServer(h)
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", orderServer)

	var handler http.Handler = mux
	// Применяем кастомные middleware
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = http.TimeoutHandler(handler, middlewareTimeout, "таймаут при обработке запроса к серверу")
	return handler, nil
}

// OrderHandler реализует интерфейс orderv1.Handler, сгенерированный ogen.
type OrderHandler struct {
	orderv1.UnimplementedHandler
	inventoryClient inventoryv1.InventoryServiceClient
	paymentClient   paymentv1.PaymentServiceClient
	store           *OrderStore
}

// NewOrderHandler создаёт новый обработчик заказов.
func NewOrderHandler(
	inventoryClient inventoryv1.InventoryServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	store *OrderStore,
) *OrderHandler {
	return &OrderHandler{
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		store:           store,
	}
}

// SetupServer создаёт OpenAPI сервер на основе обработчика.
func SetupServer(h *OrderHandler) (*orderv1.Server, error) {
	return orderv1.NewServer(h)
}

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (h *OrderHandler) GetOrder(_ context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	// 1. Найти заказ в store (с блокировкой для thread-safety)
	h.store.mu.RLock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.RUnlock()

	// 2. Если не найден — вернуть 404
	if !ok {
		return &orderv1.GetOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	// 3. Преобразовать в DTO и вернуть
	var shieldUUID orderv1.OptNilUUID
	if order.ShieldUUID != nil {
		shieldUUID = orderv1.NewOptNilUUID(*order.ShieldUUID)
	}

	var weaponUUID orderv1.OptNilUUID
	if order.WeaponUUID != nil {
		weaponUUID = orderv1.NewOptNilUUID(*order.WeaponUUID)
	}

	var transactionUUID orderv1.OptNilUUID
	if order.TransactionUUID != nil {
		transactionUUID = orderv1.NewOptNilUUID(*order.TransactionUUID)
	}

	var paymentMethod orderv1.OptNilPaymentMethod
	if order.PaymentMethod != nil {
		paymentMethod = orderv1.NewOptNilPaymentMethod(orderv1.PaymentMethod(*order.PaymentMethod))
	}

	return &orderv1.OrderDto{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      shieldUUID,
		WeaponUUID:      weaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: transactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          orderv1.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}, nil
}

// CreateOrder реализует операцию createOrder.
// POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	// 1. Валидация: hull_uuid и engine_uuid обязательны
	if err := checkRequiredField(req.GetHullUUID(), "hull_uuid"); err != nil {
		return &orderv1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}
	if err := checkRequiredField(req.GetEngineUUID(), "engine_uuid"); err != nil {
		return &orderv1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}

	// Собираю все UUID в один слайс для дальнейшей проверки на формат UUID и передачи в запрос по gRPC
	rqParts := []string{req.GetHullUUID().String(), req.GetEngineUUID().String()}
	if req.GetShieldUUID().IsSet() && !req.GetShieldUUID().IsNull() {
		rqParts = append(rqParts, req.GetShieldUUID().Value.String())
	}
	if req.GetWeaponUUID().IsSet() && !req.GetWeaponUUID().IsNull() {
		rqParts = append(rqParts, req.GetWeaponUUID().Value.String())
	}
	grpcRqParts := make([]string, len(rqParts))
	for i, p := range rqParts {
		if err := validateUUID(p); err != nil {
			return &orderv1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		}
		grpcRqParts[i] = p
	}

	// ПАТТЕРН 1: Контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	// 2. Получить детали через InventoryService.ListParts
	parts, err := h.inventoryClient.ListParts(ctx, &inventoryv1.ListPartsRequest{
		Uuids: grpcRqParts,
	})
	if err != nil {
		// ПАТТЕРН 2: Обработка gRPC ошибок
		st, msg := handleGRPCError(err)
		switch st {
		case http.StatusBadRequest:
			return &orderv1.CreateOrderBadRequest{Code: st, Message: msg}, nil
		case http.StatusNotFound:
			return &orderv1.CreateOrderNotFound{Code: st, Message: msg}, nil
		default:
			return &orderv1.CreateOrderInternalServerError{Code: st, Message: msg}, nil
		}
	}
	// Проверяет, что все детали существуют (иначе 404)
	if parts.GetParts() == nil || len(parts.GetParts()) != len(grpcRqParts) {
		msg := fmt.Sprintf("не все детали найдены: запрашивали %d, найдено %d", len(grpcRqParts), len(parts.GetParts()))
		return &orderv1.CreateOrderNotFound{Code: http.StatusNotFound, Message: msg}, nil
	}

	// 3. Проверить stock_quantity > 0
	// 4. Вычислить total_price
	var totalPrince int64
	for _, p := range parts.GetParts() {
		if p.GetStockQuantity() <= 0 {
			msg := fmt.Sprintf("количество деталей на складе для %s: %d", p.GetName(), p.GetStockQuantity())
			return &orderv1.CreateOrderConflict{Code: http.StatusConflict, Message: msg}, nil
		}
		totalPrince += p.GetPrice()
	}

	// 5. Сгенерировать order_uuid (UUID v4)
	orderUUID := uuid.New()

	// 6. Создать заказ со статусом PENDING_PAYMENT
	order := Order{
		OrderUUID:  orderUUID,
		HullUUID:   req.GetHullUUID(),
		EngineUUID: req.GetEngineUUID(),
		TotalPrice: totalPrince,
		Status:     string(pendingPayment),
		CreatedAt:  time.Now(),
	}
	if req.GetShieldUUID().IsSet() && !req.GetShieldUUID().IsNull() {
		order.ShieldUUID = &req.ShieldUUID.Value
	}
	if req.GetWeaponUUID().IsSet() && !req.GetWeaponUUID().IsNull() {
		order.WeaponUUID = &req.WeaponUUID.Value
	}

	// 7. Сохранить в store
	h.store.mu.Lock()
	h.store.orders[orderUUID] = order
	h.store.mu.Unlock()

	// 8. Вернуть order_uuid и total_price
	return &orderv1.CreateOrderResponse{
		OrderUUID:  orderUUID,
		TotalPrice: totalPrince,
	}, nil
}

// PayOrder реализует операцию payOrder.
// POST /api/v1/orders/{order_uuid}/pay.
func (h *OrderHandler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	// 1. Найти заказ в store
	h.store.mu.Lock()
	order, ok := h.store.orders[params.OrderUUID]
	defer h.store.mu.Unlock()

	if !ok {
		return &orderv1.PayOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != string(pendingPayment) {
		msg := fmt.Sprintf("статус заказа %s: %s", order.OrderUUID, order.Status)
		return &orderv1.PayOrderConflict{
			Code:    http.StatusConflict,
			Message: msg,
		}, nil
	}

	// ПАТТЕРН 1: Контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	// 3. Вызвать h.paymentClient.PayOrder для обработки платежа
	rs, err := h.paymentClient.PayOrder(ctx, &paymentv1.PayOrderRequest{
		OrderUuid:     params.OrderUUID.String(),
		PaymentMethod: mapGRPCPayment(req.GetPaymentMethod()),
	})
	if err != nil {
		// ПАТТЕРН 2: Обработка gRPC ошибок
		st, msg := handleGRPCError(err)
		switch st {
		case http.StatusBadRequest:
			return &orderv1.PayOrderBadRequest{Code: st, Message: msg}, nil
		case http.StatusNotFound:
			return &orderv1.PayOrderNotFound{Code: st, Message: msg}, nil
		default:
			return &orderv1.PayOrderInternalServerError{Code: st, Message: msg}, nil
		}
	}

	// 4. Обновить статус на PAID и сохранить transaction_uuid
	// и payment_method
	transactionUUID, err := uuid.Parse(rs.GetTransactionUuid())
	if err != nil {
		msg := fmt.Sprintf("неверный формат uuid: %s", transactionUUID)
		return &orderv1.PayOrderConflict{
			Code:    http.StatusConflict,
			Message: msg,
		}, nil
	}
	order.TransactionUUID = &transactionUUID
	order.PaymentMethod = new(string(req.PaymentMethod))
	order.Status = string(paid)
	h.store.orders[order.OrderUUID] = order

	// 5. Вернуть transaction_uuid
	return &orderv1.PayOrderResponse{TransactionUUID: transactionUUID}, nil
}

// CancelOrder реализует операцию cancelOrder.
// POST /api/v1/orders/{order_uuid}/cancel.
func (h *OrderHandler) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	// 1. Найти заказ в store
	h.store.mu.Lock()
	order, ok := h.store.orders[params.OrderUUID]
	defer h.store.mu.Unlock()

	if !ok {
		return &orderv1.CancelOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != string(pendingPayment) {
		msg := fmt.Sprint("невозможно отменить заказ в статусе ", order.Status)
		return &orderv1.CancelOrderConflict{
			Code:    http.StatusConflict,
			Message: msg,
		}, nil
	}

	// 3. Обновить статус на CANCELLED
	order.Status = string(cancelled)
	h.store.orders[params.OrderUUID] = order

	// 4. Вернуть success
	return &orderv1.CancelOrderResponse{}, nil
}

func checkRequiredField(field uuid.UUID, fieldName string) error {
	if field == uuid.Nil {
		return fmt.Errorf("в запросе не передано обязательное поле: %s", fieldName)
	}
	return nil
}

func validateUUID(val string) error {
	if _, err := uuid.Parse(val); err != nil {
		return fmt.Errorf("неверный формат uuid: %s", val)
	}
	return nil
}

// handleGRPCError конвертирует gRPC ошибки в HTTP ответы.
// Это КЛЮЧЕВОЙ ПАТТЕРН: правильная конвертация кодов ошибок.
func handleGRPCError(err error) (code int, message string) {
	st, ok := status.FromError(err)
	if !ok {
		// Неизвестная ошибка
		slog.Error("неизвестная ошибка", "error", err)
		return http.StatusInternalServerError, "внутренняя ошибка сервера"
	}
	// Маппинг gRPC кодов → HTTP кодов
	switch st.Code() {
	case codes.NotFound:
		return http.StatusNotFound, st.Message()
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.DeadlineExceeded: // Таймаут — клиент может повторить запрос
		return http.StatusGatewayTimeout, "таймаут запроса"
	case codes.Unavailable: // gRPC сервис недоступен
		return http.StatusServiceUnavailable, "бэкенд сервис недоступен"
	case codes.Canceled: // Запрос был отменён (например, клиент закрыл соединение)
		slog.Info("запрос отменен", "message", st.Message())
		return http.StatusServiceUnavailable, "запрос отменен"
	default:
		slog.Info("ошибка gRPC", "code", st.Code(), "message", st.Message())
		return http.StatusInternalServerError, "внутренняя ошибка сервера"
	}
}

func mapGRPCPayment(p orderv1.PaymentMethod) paymentv1.PaymentMethod {
	switch p {
	case orderv1.PaymentMethodCARD:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CARD
	case orderv1.PaymentMethodSBP:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_SBP
	case orderv1.PaymentMethodCREDITCARD:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case orderv1.PaymentMethodINVESTORMONEY:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
