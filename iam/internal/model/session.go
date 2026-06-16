package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UUID      uuid.UUID
	CreatedAt time.Time
	UpdatedAt *time.Time
	ExpiresAt time.Time
}

func NewSession(ttl time.Duration) Session {
	now := time.Now()
	return Session{
		UUID:      uuid.New(),
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}
