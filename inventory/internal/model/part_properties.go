package model

// PartProperties — экспортируемый алиас для приватной структуры.
type PartProperties = partProperties

// partProperties — типоспецифичные свойства детали.
// Ровно одно поле non-nil — определяется типом детали.
type partProperties struct {
	hull   *HullProperties
	engine *EngineProperties
	shield *ShieldProperties
	weapon *WeaponProperties
}

// Hull возвращает свойства корпуса или nil.
func (p *partProperties) Hull() *HullProperties { return p.hull }

// Engine возвращает свойства двигателя или nil.
func (p *partProperties) Engine() *EngineProperties { return p.engine }

// Shield возвращает свойства защиты или nil.
func (p *partProperties) Shield() *ShieldProperties { return p.shield }

// Weapon возвращает свойства вооружения или nil.
func (p *partProperties) Weapon() *WeaponProperties { return p.weapon }
