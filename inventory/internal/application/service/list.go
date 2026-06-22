package part

import (
	"context"
	"errors"
	"slices"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/inventory/internal/application/input"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) List(ctx context.Context, filter input.PartFilter) ([]model.Part, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.List", trace.WithAttributes(
		attribute.StringSlice("item.uuids", uuidsToStringSlice(filter.UUIDs)),
		attribute.String("item.type", string(filter.PartType)),
	))
	defer span.End()

	if filter.UUIDs != nil {
		partsUnordered, err := s.partRepository.ListByUUIDs(ctx, filter.UUIDs)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			switch {
			case errors.Is(err, errs.ErrPartNotFound):
				return nil, sharedErrs.NewNotFound(err)
			case errors.Is(err, errs.ErrWrongPartType):
				return nil, sharedErrs.NewConflict(err)
			default:
				return nil, sharedErrs.NewInternal(err)
			}
		}

		return model.OrderPartsByIds(partsUnordered, filter.UUIDs), nil
	}

	var (
		parts []model.Part
		err   error
	)

	// Если part_type == UNSPECIFIED → вернуть все детали
	if filter.PartType == "" {
		parts, err = s.partRepository.List(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return nil, sharedErrs.NewInternal(err)
		}
	} else { // Иначе → фильтровать по типу
		parts, err = s.partRepository.ListByPartType(ctx, filter.PartType)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return nil, sharedErrs.NewInternal(err)
		}
	}

	// Отсортировать по имени (только для фильтрации по типу, не для uuids)
	if parts != nil {
		slices.SortFunc(parts, func(a, b model.Part) int {
			return strings.Compare(a.Name(), b.Name())
		})
	}

	span.SetAttributes(attribute.Int("items.count", len(parts)))

	return parts, nil
}
