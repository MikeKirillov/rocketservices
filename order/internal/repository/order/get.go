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
	query := `SELECT
	o.uuid, o.status, o.transaction_uuid, o.payment_method,
	o.created_at, o.updated_at, o.version, o.user_uuid,
	COALESCE(
	    jsonb_agg(
	        jsonb_build_object(
	            'uuid',       oi.uuid,
				'order_uuid', oi.order_uuid,
	            'part_uuid',  oi.part_uuid,
	            'part_type',  oi.part_type,
	            'price',      oi.price,
	            'created_at', to_char(oi.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
	        ) ORDER BY oi.created_at
	    ) FILTER (WHERE oi.uuid IS NOT NULL),
	    '[]'::jsonb
	) AS items
	FROM orders o
	LEFT JOIN order_items oi ON oi.order_uuid = o.uuid
	WHERE o.uuid = $1
	GROUP BY o.uuid`
	// FILTER (WHERE oi.uuid IS NOT NULL) - чтобы при заказе без позиций не получить [null].
	// COALESCE(..., '[]') - пустой массив вместо NULL.

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return model.Order{}, err
	}
	defer rows.Close()

	order, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[record.Order])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, errs.ErrOrderNotFound
		}
		slog.Info("не удалось прочитать строку", "error", err)
		return model.Order{}, err
	}

	if rows.Err() != nil {
		slog.Error("ошибка при итерации по строкам", "error", err)
		return model.Order{}, err
	}

	return converter.OrderToModel(order), nil
}
