package order

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	query := `SELECT
	o.uuid, o.status, o.transaction_uuid, o.payment_method, o.created_at, o.updated_at,
	i.uuid as i_uuid, i.order_uuid, i.part_uuid, i.part_type, i.price, i.created_at as i_created_at
	FROM orders o
	JOIN order_items i on o.uuid = i.order_uuid
	WHERE o.uuid = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return model.Order{}, err
	}
	defer rows.Close()

	order := record.Order{}
	var items []record.OrderItem

	found := false
	for rows.Next() {
		found = true

		i := record.OrderItem{}
		err = rows.Scan(
			&order.UUID, &order.Status, &order.TransactionUUID, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt,
			&i.UUID, &i.OrderUUID, &i.PartUUID, &i.PartType, &i.Price, &i.CreatedAt,
		)
		if err != nil {
			slog.Info("не удалось прочитать строку", "error", err)
			return model.Order{}, err
		}

		// проверяем, что LEFT JOIN притянул записи из таблицы order_items
		if i.UUID != uuid.Nil {
			items = append(items, i)
		}
	}

	if !found {
		return model.Order{}, errs.ErrOrderNotFound
	}

	if rows.Err() != nil {
		slog.Error("ошибка при итерации по строкам", "error", err)
		return model.Order{}, err
	}
	return converter.OrderToModel(order, items), nil
}
