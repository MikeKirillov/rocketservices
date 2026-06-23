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
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "inventory.List.Ok")
		}
		span.End()
	}()

	if filter.UUIDs != nil {
		partsUnordered, err := s.partRepository.ListByUUIDs(ctx, filter.UUIDs)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrPartNotFound):
				{
					resultErr = sharedErrs.NewNotFound(err)
					return nil, resultErr
				}
			case errors.Is(err, errs.ErrWrongPartType):
				{
					resultErr = sharedErrs.NewConflict(err)
					return nil, resultErr
				}
			default:
				{
					resultErr = sharedErrs.NewInternal(err)
					return nil, resultErr
				}
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
			resultErr = sharedErrs.NewInternal(err)
			return nil, resultErr
		}
	} else { // Иначе → фильтровать по типу
		parts, err = s.partRepository.ListByPartType(ctx, filter.PartType)
		if err != nil {
			resultErr = sharedErrs.NewInternal(err)
			return nil, resultErr
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
