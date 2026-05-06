package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return model.Order{}, sharedErrs.NewNotFound(err)
		}
		return model.Order{}, sharedErrs.NewInternal(err)
	}
	return order, nil
}
