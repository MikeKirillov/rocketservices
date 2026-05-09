package order

import (
	"context"
	"fmt"
	"time"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

func (r *repository) Update(ctx context.Context, order model.Order) error {
	return r.txManager.Do(ctx, func(txCtx context.Context) error {
		query := `UPDATE orders SET
		status = $1,
		transaction_uuid = COALESCE($2, transaction_uuid),
		payment_method = COALESCE($3, payment_method),
		updated_at = $4
		WHERE uuid = $5`

		db := r.getter.DefaultTrOrDB(txCtx, r.pool)
		tag, err := db.Exec(
			txCtx, query, order.Status, order.TransactionUUID, order.PaymentMethod, time.Now(), order.OrderUUID,
		)
		if err != nil {
			return fmt.Errorf("обновление записи %s: %w", order.OrderUUID, err)
		}

		if tag.RowsAffected() == 0 {
			return errs.ErrOrderNotFound
		}

		return nil
	})
}
