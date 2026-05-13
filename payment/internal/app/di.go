package app

import (
	"context"

	paymentAPIV1 "github.com/MikeKirillov/rocketservices/payment/internal/api/payment/v1"
	orderService "github.com/MikeKirillov/rocketservices/payment/internal/service/payment"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

// DI-контейнер с ленивой инициализацией зависимостей через геттеры.
type diContainer struct {
	paymentSvc       paymentAPIV1.OrderService
	paymentV1Handler paymentv1.PaymentServiceServer
}

// Service возвращает сервис бизнес-логики оплаты заказа.
func (d *diContainer) PaymentService(_ context.Context) paymentAPIV1.OrderService {
	if d.paymentSvc == nil {
		d.paymentSvc = orderService.NewService()
	}
	return d.paymentSvc
}

// Handler возвращает gRPC-обработчик сервиса наблюдений НЛО.
func (d *diContainer) PaymentV1API(ctx context.Context) paymentv1.PaymentServiceServer {
	if d.paymentV1Handler == nil {
		d.paymentV1Handler = paymentAPIV1.NewAPI(d.PaymentService(ctx))
	}
	return d.paymentV1Handler
}
