package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderPaidEvent struct {
	OrderUuid uuid.UUID
	UserUuid  uuid.UUID
	Received  time.Time
}
