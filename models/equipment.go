package models

type Equipment struct {
	Weapon         int      `json:"weapon" firestore:"weapon"`
	Body           int      `json:"body" firestore:"body"`
	Glove          int      `json:"glove" firestore:"glove"`
	Shoes          int      `json:"shoes" firestore:"shoes"`
	Bindi          *int     `json:"bindi" firestore:"bindi,omitempty"`
	Glasses        *int     `json:"glasses" firestore:"glasses,omitempty"`
	Earring        *int     `json:"earrings" firestore:"earrings,omitempty"`
	Ring           *int     `json:"ring" firestore:"ring,omitempty"`
	Mantle         *int     `json:"mantle" firestore:"mantle,omitempty"`
	Stockings      *int     `json:"stockings" firestore:"stockings,omitempty"`
	EquipmentNames []string `json:"equipmentNames,omitempty" firestore:"equipmentNames,omitempty"`
}

type EquipmentSheet struct {
	Name                string            `json:"name" firestore:"name,omitempty"`
	ID                  string            `json:"id" firestore:"id,omitempty"`
	Cost                int32             `json:"cost" firestore:"cost,omitempty"`
	AccessoryCost       int32             `json:"accessoryCost" firestore:"accessoryCost"`
	LevelRequirement    int32             `json:"levelRequirement" firestore:"levelRequirement,omitempty"`
	TierRequirement     int32             `json:"tierRequirement" firestore:"tierRequirement"`
	ShoeEvasion         float64           `json:"shoeEvasion" firestore:"shoeEvasion"`
	GloveAccuracy       float64           `json:"gloveAccuracy" firestore:"gloveAccuracy"`
	GloveCriticalDamage float64           `json:"gloveCriticalDamage" firestore:"gloveCriticalDamage"`
	BindiHP             float64           `json:"bindiHP" firestore:"bindiHP"`
	GlassesCritDamage   float64           `json:"glassesCritDamage" firestore:"glassesCritDamage"`
	EarringCritRate     float64           `json:"earringCritRate" firestore:"earringCritRate"`
	RingCritRate        float64           `json:"ringCritRate" firestore:"ringCritRate"`
	MantleDamage        float64           `json:"mantleDamage" firestore:"mantleDamage"`
	StockingEvasion     float64           `json:"stockingEvasion" firestore:"stockingEvasion"`
	ArmorDefense        float64           `json:"armorDefense" firestore:"armorDefense"`
	WeaponDPS           float64           `json:"weaponDPS" firestore:"weaponDPS"`
	WeaponMap           map[string]string `json:"weapon" firestore:"weapon,omitempty"`
	WeaponList          []WeaponType
}

type WeaponType struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}
