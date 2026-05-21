package part

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/MikeKirillov/rocketservices/inventory/internal/application/input"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) List(ctx context.Context, filter input.PartFilter) ([]model.Part, error) {
	if filter.UUIDs != nil {
		partsUnordered, err := s.partRepository.ListByUUIDs(ctx, filter.UUIDs)
		if err != nil {
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
			return nil, sharedErrs.NewInternal(err)
		}
	} else { // Иначе → фильтровать по типу
		parts, err = s.partRepository.ListByPartType(ctx, filter.PartType)
		if err != nil {
			return nil, sharedErrs.NewInternal(err)
		}
	}

	// Отсортировать по имени (только для фильтрации по типу, не для uuids)
	if parts != nil {
		slices.SortFunc(parts, func(a, b model.Part) int {
			return strings.Compare(a.Name(), b.Name())
		})
	}
	return parts, nil
}
