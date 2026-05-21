package part

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

type repository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool:   pool,
		getter: trmpgx.DefaultCtxGetter,
	}
}

func (r *repository) Get(ctx context.Context, uuid uuid.UUID) (model.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE uuid = $1`

	var p record.Part
	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(ctx, query, uuid).Scan(
		&p.UUID,
		&p.Name,
		&p.Description,
		&p.PartType,
		&p.Price,
		&p.StockQuantity,
		&p.Reserved,
		&p.Properties,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Part{}, errs.ErrPartNotFound
		}
		return model.Part{}, err
	}
	part, err := converter.PartRecordToModel(p)
	if err != nil {
		return model.Part{}, err
	}
	return part, nil
}

func (r *repository) List(ctx context.Context) ([]model.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts`

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query)
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return nil, err
	}
	defer rows.Close()

	p, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	parts, err := converter.PartsToModel(p)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

func (r *repository) ListByPartType(ctx context.Context, partType model.PartType) ([]model.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE part_type = $1`
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, partType)
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return nil, err
	}
	defer rows.Close()

	p, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	parts, err := converter.PartsToModel(p)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

func (r *repository) ListByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE uuid = ANY($1)`

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, uuids)
	if err != nil {
		slog.Error("не удалось выполнить выборку", "error", err)
		return nil, err
	}
	defer rows.Close()

	p, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Создаем set параметров UUID из ответа БД
	foundUUIDs := make(map[uuid.UUID]struct{})
	for _, part := range p {
		foundUUIDs[part.UUID] = struct{}{}
	}

	// Проверяем, какие UUID из запроса не были найдены
	var missingUUIDs []uuid.UUID
	for _, id := range uuids {
		if _, ok := foundUUIDs[id]; !ok {
			missingUUIDs = append(missingUUIDs, id)
		}
	}

	if len(missingUUIDs) > 0 {
		return nil, fmt.Errorf("детали %s: %w", missingUUIDs, errs.ErrPartNotFound)
	}

	parts, err := converter.PartsToModel(p)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

func scanRows(rows pgx.Rows) ([]record.Part, error) {
	parts, err := pgx.CollectRows(rows, pgx.RowToStructByName[record.Part])
	if err != nil {
		slog.Info("не удалось прочитать строки", "error", err)
		return nil, err
	}
	return parts, nil
}

func (r *repository) UpdateReservedBatch(ctx context.Context, parts []model.Part) error {
	query := `UPDATE parts AS p
	SET reserved = batch.reserved
	FROM unnest($1::uuid[], $2::int[]) AS batch(uuid, reserved)
	WHERE p.uuid = batch.uuid`

	uuids := make([]uuid.UUID, len(parts))
	reservedVal := make([]int, len(parts))

	for i, p := range parts {
		uuids[i] = p.UUID()
		reservedVal[i] = p.Reserved()
	}

	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, uuids, reservedVal)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("не удалось обновить записи")
	}

	return nil
}
