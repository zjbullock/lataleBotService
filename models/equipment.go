package models

type Equipment struct {
	Weapon         int      `json:"weapon" firestore:"weapon"`
	Body           int      `json:"body" firestore:"body"`
	Glove          int      `json:"glove" firestore:"glove"`
	Shoes          int      `json:"shoes" firestore:"shoes"`
	EquipmentNames []string `json:"equipmentNames,omitempty" firestore:"equipmentNames,omitempty"`
}

type EquipmentSheet struct {
	Name             string            `json:"name" firestore:"name,omitempty"`
	ID               string            `json:"id" firestore:"id,omitempty"`
	Cost             int32             `json:"cost" firestore:"cost,omitempty"`
	LevelRequirement int32             `json:"levelRequirement" firestore:"levelRequirement,omitempty"`
	ShoeEvasion      float64           `json:"shoeEvasion" firestore:"shoeEvasion"`
	GloveAccuracy    float64           `json:"gloveAccuracy" firestore:"gloveAccuracy"`
	ArmorDefense     float64           `json:"armorDefense" firestore:"armorDefense"`
	WeaponDPS        float64           `json:"weaponDPS" firestore:"weaponDPS"`
	WeaponMap        map[string]string `json:"weapon" firestore:"weapon,omitempty"`
	WeaponList       []WeaponType
}

type WeaponType struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}
