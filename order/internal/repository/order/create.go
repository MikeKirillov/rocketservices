package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
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
	o := converter.OrderToRecord(order)

	query := `INSERT INTO orders (uuid, status, created_at) VALUES ($1, $2, $3)`

	db := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, o.UUID, o.Status, o.CreatedAt)
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
	for _, item := range items {
		i := converter.OrderItemToRecord(uuid.New(), order.OrderUUID, item)
		batch.Queue(query, i.UUID, i.OrderUUID, i.PartUUID, i.PartType, i.Price)
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
