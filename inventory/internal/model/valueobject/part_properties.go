package valueobject

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
func (p *PartProperties) Hull() *HullProperties { return p.hull }

// Engine возвращает свойства двигателя или nil.
func (p *PartProperties) Engine() *EngineProperties { return p.engine }

// Shield возвращает свойства защиты или nil.
func (p *PartProperties) Shield() *ShieldProperties { return p.shield }

// Weapon возвращает свойства вооружения или nil.
func (p *PartProperties) Weapon() *WeaponProperties { return p.weapon }
