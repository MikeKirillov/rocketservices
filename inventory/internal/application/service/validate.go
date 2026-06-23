package part

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/inventory/internal/application/input"
	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Validate(ctx context.Context, slots input.ShipSlots) error {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Validate", trace.WithAttributes(
		attribute.String("hull.uuid", slots.HullUUID.String()),
		attribute.String("engine.uuid", slots.EngineUUID.String()),
		attribute.String("shield.uuid", slots.ShieldUUID.String()),
		attribute.String("weapon.uuid", slots.WeaponUUID.String()),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "inventory.Validate.Ok")
		}
		span.End()
	}()

	// Запретите использование одного UUID в разных слотах → тоже ErrPartTypeMismatch.
	seen := make(map[uuid.UUID]bool)
	for _, id := range []uuid.UUID{slots.HullUUID, slots.EngineUUID, slots.ShieldUUID, slots.WeaponUUID} {
		if id == uuid.Nil {
			continue
		}
		if seen[id] {
			return sharedErrs.NewInvalidArgument(errs.ErrPartTypeMismatch)
		}
		seen[id] = true
	}

	// Загрузите детали по непустым UUID одним вызовом partRepo.List.
	ids := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	parts, err := s.partRepository.ListByUUIDs(ctx, ids)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPartNotFound):
			{
				resultErr = sharedErrs.NewNotFound(err)
				return resultErr
			}
		case errors.Is(err, errs.ErrWrongPartType):
			{
				resultErr = sharedErrs.NewConflict(err)
				return resultErr
			}
		default:
			{
				resultErr = sharedErrs.NewInternal(err)
				return resultErr
			}
		}
	}

	// Для каждого слота проверьте, что вернувшаяся деталь имеет ожидаемый тип. Несоответствие → ErrPartTypeMismatch.
	pMap := make(map[uuid.UUID]model.Part, len(parts))
	for _, p := range parts {
		pMap[p.UUID()] = p
	}

	resolved, err := resolvedShipSlots(pMap, slots)
	if err != nil {
		resultErr = sharedErrs.NewInvalidArgument(err)
		return resultErr
	}

	// Соберите проверенные детали в model.ResolvedShipSlots и передайте в domain service.
	err = s.compatibilityChecker.Check(resolved)
	if err != nil {
		resultErr = sharedErrs.NewFailedPrecondition(err)
		return resultErr
	}
	return nil
}

// resolvedShipSlots проверяет каждую деталь, полученную из БД, на соответствие типа
// детали согласно запрошенным слотам.
func resolvedShipSlots(pMap map[uuid.UUID]model.Part, slots input.ShipSlots) (model.ResolvedShipSlots, error) {
	resolved := model.ResolvedShipSlots{}

	if p, ok := pMap[slots.HullUUID]; ok && p.PartType() != model.PartTypeHull {
		return model.ResolvedShipSlots{}, errs.ErrPartTypeMismatch
	} else {
		resolved.Hull = p
	}

	if p, ok := pMap[slots.EngineUUID]; ok && p.PartType() != model.PartTypeEngine {
		return model.ResolvedShipSlots{}, errs.ErrPartTypeMismatch
	} else {
		resolved.Engine = p
	}

	if slots.ShieldUUID != uuid.Nil {
		p := pMap[slots.ShieldUUID]
		if p.PartType() != model.PartTypeShield {
			return model.ResolvedShipSlots{}, errs.ErrPartTypeMismatch
		} else {
			resolved.Shield = &p
		}
	}

	if slots.WeaponUUID != uuid.Nil {
		p := pMap[slots.WeaponUUID]
		if p.PartType() != model.PartTypeWeapon {
			return model.ResolvedShipSlots{}, errs.ErrPartTypeMismatch
		} else {
			resolved.Weapon = &p
		}
	}

	return resolved, nil
}
