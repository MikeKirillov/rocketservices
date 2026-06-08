package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UUID      uuid.UUID
	UserUUID  uuid.UUID
	Login     string
	CreatedAt time.Time
	UpdatedAt *time.Time
	ExpiresAt time.Time
}
