package order

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	queryOrder := `SELECT
	uuid, status, transaction_uuid, payment_method, created_at, updated_at
	FROM public.orders
	WHERE uuid = $1`

	queryItems := `SELECT
    uuid, order_uuid, part_uuid, part_type, price, created_at
	FROM public.order_items
	WHERE order_uuid = $1`

	batch := new(pgx.Batch{})
	batch.Queue(queryOrder, id)
	batch.Queue(queryItems, id)

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var o record.Order
	err := br.QueryRow().Scan(&o.UUID, &o.Status, &o.TransactionUUID, &o.PaymentMethod, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, errs.ErrOrderNotFound
		}
		return model.Order{}, err
	}

	rows, err := br.Query()
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return model.Order{}, err
	}
	defer rows.Close()

	var items []record.OrderItem
	for rows.Next() {
		var i record.OrderItem

		err = rows.Scan(&i.UUID, &i.OrderUUID, &i.PartUUID, &i.PartType, &i.Price, &i.CreatedAt)
		if err != nil {
			slog.Info("не удалось прочитать строку", "error", err)
			return model.Order{}, err
		}

		items = append(items, i)
	}

	if err = rows.Err(); err != nil {
		slog.Error("ошибка при итерации по строкам", "error", err)
		return model.Order{}, err
	}

	return converter.OrderToModel(o, items), nil
}
