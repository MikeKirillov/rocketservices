package record

import (
	"bytes"
	"strings"
	"time"

	"github.com/google/uuid"
)

const layout = "2006-01-02T15:04:05.000000"

type JSONTime time.Time

func (jt *JSONTime) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) || len(b) == 0 {
		return nil
	}

	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}

	t, err := time.Parse(layout, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
	}

	*jt = JSONTime(t)
	return nil
}

type Order struct {
	UUID            uuid.UUID   `db:"uuid"`
	Items           []OrderItem `db:"items"`
	Status          string      `db:"status"`
	TransactionUUID *uuid.UUID  `db:"transaction_uuid"` // nullable → указатель
	PaymentMethod   *string     `db:"payment_method"`
	CreatedAt       time.Time   `db:"created_at"`
	UpdatedAt       *time.Time  `db:"updated_at"`
	Version         int64       `db:"version"`
}

type OrderItem struct {
	UUID         uuid.UUID `json:"uuid"` // json:"..." для декодинга с применением jsonb
	OrderUUID    uuid.UUID `json:"order_uuid"`
	PartUUID     uuid.UUID `json:"part_uuid"`
	PartType     string    `json:"part_type"`
	Price        int64     `json:"price"`
	CreatedAtRaw JSONTime  `json:"created_at"` // строка JSON для JSONB - используется при поиске записей в БД
	CreatedAt    time.Time
}
