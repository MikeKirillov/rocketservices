package order

import (
	"context"
	"sync"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

type repository struct {
	mu   sync.RWMutex
	data map[string]record.Order
}

func NewRepository() *repository {
	return &repository{data: make(map[string]record.Order)}
}

func (r *repository) Get(_ context.Context, id string) (model.Order, error) {
	// Найти заказ (с блокировкой для thread-safety)
	r.mu.RLock()
	order, ok := r.data[id]
	r.mu.RUnlock()

	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	return converter.OrderToModel(order), nil
}

func (r *repository) Create(_ context.Context, order record.Order) (string, error) {
	r.mu.Lock()
	r.data[order.OrderUUID] = order
	r.mu.Unlock()

	return order.OrderUUID, nil
}

func (r *repository) Update(_ context.Context, order record.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[order.OrderUUID]; !ok {
		return errs.ErrOrderNotFound
	}

	r.data[order.OrderUUID] = order
	return nil
}
