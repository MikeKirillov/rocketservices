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

func NewSession(user User) Session {
	return Session{
		UUID:      uuid.New(),
		UserUUID:  user.UUID,
		Login:     user.Login,
		CreatedAt: time.Now(),
	}
}
