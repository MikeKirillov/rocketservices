package model

import (
	"time"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
)

// Part — экспортируемый алиас для приватной структуры.
type Part = part

// part представляет деталь космического корабля.
type part struct {
	uuid          uuid.UUID
	name          string
	description   string
	price         int64 // в копейках
	partType      PartType
	stockQuantity int
	reserved      int
	properties    PartProperties
	createdAt     time.Time
}

// UUID возвращает уникальный идентификатор детали.
func (p *part) UUID() uuid.UUID { return p.uuid }

// Name возвращает наименование детали.
func (p *part) Name() string { return p.name }

// Description возвращает описание детали.
func (p *part) Description() string { return p.description }

// Price возвращает стоимость одной детали.
func (p *part) Price() int64 { return p.price }

// PartType возвращает тип детали.
func (p *part) PartType() PartType { return p.partType }

// StockQuantity возвращает остаток деталей на складе.
func (p *part) StockQuantity() int { return p.stockQuantity }

// Reserved возвращает количество зарезервированных деталей.
func (p *part) Reserved() int { return p.reserved }

// Properties возвращает свойства детали.
func (p *part) Properties() PartProperties { return p.properties }

// CreatedAt возвращает дату создания записи детали в БД.
func (p *part) CreatedAt() time.Time { return p.createdAt }

// Reserve резервирует одну единицу компонента.
// Не может превысить stockQuantity и не может стать отрицательным.
// Возвращает ErrOutOfStock, если свободных единиц нет.
func (p *part) Reserve() error {
	if p.reserved >= p.stockQuantity {
		return errs.ErrOutOfStock
	}
	p.reserved++
	return nil
}

// Release освобождает одну зарезервированную единицу компонента.
// Возвращает ErrNothingToRelease, если резерв равен нулю.
func (p *part) Release() error {
	if p.reserved <= 0 {
		return errs.ErrNothingToRelease
	}
	p.reserved--
	return nil
}

// Commit освобождает одну зарезервированную единицу компонента.
// Возвращает ErrNothingToCommit, если резерв равен нулю.
func (p *Part) Commit() error {
	if p.reserved == 0 || p.stockQuantity == 0 {
		return errs.ErrNothingToCommit
	}
	p.reserved--
	p.stockQuantity--
	return nil
}

// RestorePart восстанавливает сущность из БД (без валидации — данные уже проверены).
func RestorePart(
	id uuid.UUID,
	name, description string,
	price int64,
	partType PartType,
	stockQuantity, reserved int,
	properties PartProperties,
	createdAt time.Time,
) Part {
	return part{
		uuid:          id,
		name:          name,
		description:   description,
		price:         price,
		partType:      partType,
		stockQuantity: stockQuantity,
		reserved:      reserved,
		properties:    properties,
		createdAt:     createdAt,
	}
}

// OrderPartsByIds сортирует список деталей в соответствии со списком идентификаторов.
func OrderPartsByIds(parts []Part, ids []uuid.UUID) []Part {
	m := make(map[uuid.UUID]Part)
	for _, p := range parts {
		m[p.UUID()] = p
	}

	ordered := make([]Part, 0, len(ids))
	for _, id := range ids {
		if p, ok := m[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered
}
