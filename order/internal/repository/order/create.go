package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

func (r *repository) Create(ctx context.Context, order model.Order) error {
	return r.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := r.createOrder(txCtx, order); err != nil {
			return err
		}
		return r.createOrderItem(txCtx, order)
	})
}

func (r *repository) createOrder(ctx context.Context, order model.Order) error {
	query := `INSERT INTO orders (uuid, status, created_at) VALUES ($1, $2, $3)`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, order.OrderUUID, string(order.Status), order.CreatedAt)
	if err != nil {
		return fmt.Errorf("создание записи заказа %s: %w", order.OrderUUID, err)
	}

	return nil
}

func (r *repository) createOrderItem(ctx context.Context, order model.Order) error {
	items := order.Items

	if len(items) == 0 {
		return nil
	}

	query := `INSERT INTO order_items (uuid, order_uuid, part_uuid, part_type, price) 
	VALUES ($1, $2, $3, $4, $5)`

	batch := new(pgx.Batch{})
	for _, i := range items {
		batch.Queue(query, uuid.New(), order.OrderUUID, i.Uuid, string(i.Type), i.Price)
	}

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	br := db.SendBatch(ctx, batch)
	defer br.Close()

	for i := range items {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("ошибка создания записи %s: %w", items[i].Uuid, err)
		}
	}

	return nil
}
