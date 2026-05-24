package order

import (
	"context"
	"fmt"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, order model.Order) error {
	o := converter.OrderToRecord(order)

	query := `UPDATE orders SET
		status = $1,
		transaction_uuid = $2,
		payment_method = $3,
		updated_at = $4,
		version = version + 1,
		user_uuid = $5
		WHERE uuid = $6 AND version = $7`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	tag, err := db.Exec(
		ctx, query, o.Status, o.TransactionUUID, o.PaymentMethod, o.UpdatedAt, o.UserUUID, o.UUID, o.Version,
	)
	if err != nil {
		return fmt.Errorf("обновление записи %s: %w", o.UUID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("обновление записи %s (версия %d): %w", o.UUID, o.Version, errs.ErrOrderConflict)
	}

	return nil
}
