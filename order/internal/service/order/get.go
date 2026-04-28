package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

func (s *service) Get(ctx context.Context, id string) (model.Order, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		slog.Error("заказ не найден", "uuid", id)
		return model.Order{}, fmt.Errorf("получить заказ: %w", err)
	}
	return order, nil
}
