package order

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("order-service")

var (
	ordersCreated, _ = meter.Int64Counter("orders_created", metric.WithDescription("Количество созданных заказов"))
	ordersPaid, _    = meter.Int64Counter("orders_paid", metric.WithDescription("Количество оплаченных заказов"))
	ordersRevenue, _ = meter.Int64Counter("orders_revenue", metric.WithDescription("Суммарная выручка (в копейках)"))
)
