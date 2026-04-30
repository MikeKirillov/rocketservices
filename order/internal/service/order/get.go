package order

import (
	"context"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id string) (model.Order, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return model.Order{}, sharedErrs.NewNotFound(err)
	}
	return order, nil
}
