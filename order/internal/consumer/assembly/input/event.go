package input

import (
	"time"

	"github.com/google/uuid"
)

type ShipAssembledEvent struct {
	OrderUuid    uuid.UUID
	UserUuid     uuid.UUID
	BuildTimeSec int64
	AssembledAt  time.Time
}
