package input

import (
	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

type CreateOrderRq struct {
	Items    []model.OrderItem
	UserUUID uuid.UUID
}

func (r *CreateOrderRq) PartUUIDsAsSlice() []uuid.UUID {
	uuids := make([]uuid.UUID, 0, len(r.Items))
	for _, v := range r.Items {
		uuids = append(uuids, v.Uuid)
	}
	return uuids
}

type CreateOrderRs struct {
	OrderUUID  uuid.UUID
	TotalPrice int64 // в копейках
}

type PayOrderRq struct {
	OrderUUID     uuid.UUID
	PaymentMethod model.PaymentMethod
}
