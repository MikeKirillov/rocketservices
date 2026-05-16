package part

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/valueobject"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/converter"
	"github.com/MikeKirillov/rocketservices/inventory/internal/repository/record"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

func (r *repository) Get(ctx context.Context, uuid uuid.UUID) (entity.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE uuid = $1`

	var p record.Part
	err := r.pool.QueryRow(ctx, query, uuid).Scan(
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
			return entity.Part{}, errs.ErrPartNotFound
		}
		return entity.Part{}, err
	}
	part, err := converter.PartRecordToModel(p)
	if err != nil {
		return entity.Part{}, err
	}
	return part, nil
}

func (r *repository) List(ctx context.Context) ([]entity.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts`

	rows, err := r.pool.Query(ctx, query)
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

func (r *repository) ListByPartType(ctx context.Context, partType valueobject.PartType) ([]entity.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE part_type = $1`
	rows, err := r.pool.Query(ctx, query, partType)
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

func (r *repository) ListByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]entity.Part, error) {
	query := `SELECT
	uuid, name, description, part_type, price, stock_quantity, reserved, properties, created_at
	FROM parts WHERE uuid = ANY($1)`

	rows, err := r.pool.Query(ctx, query, uuids)
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
