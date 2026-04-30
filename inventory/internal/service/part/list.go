package part

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

const partTypeUnspecified = "UNSPECIFIED"

func (s *service) List(ctx context.Context, filter model.PartFilter) ([]model.Part, error) {
	if filter.UUIDs != nil {
		// Проверить формат каждого UUID → INVALID_ARGUMENT
		for _, id := range filter.UUIDs {
			if _, err := uuid.Parse(id); err != nil {
				return nil, sharedErrs.NewInvalidArgument(fmt.Errorf("деталь %s: %w", id, errs.ErrPartInvalidUUID))
			}
		}

		parts, err := s.partRepository.ListByUUIDs(ctx, filter.UUIDs)
		if err != nil {
			return nil, sharedErrs.NewNotFound(err)
		}
		return parts, nil
	}

	var (
		parts []model.Part
		err   error
	)
	// Если part_type == UNSPECIFIED → вернуть все детали
	if filter.PartType == partTypeUnspecified {
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
			return strings.Compare(a.Name, b.Name)
		})
	}
	return parts, nil
}
