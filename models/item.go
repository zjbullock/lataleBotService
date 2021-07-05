package models

type Item struct {
	Name             string        `json:"name" firestore:"name"`
	Type             ItemType      `json:"type" firestore:"type"`
	LevelRequirement *float64      `json:"levelRequirement,omitempty" firestore:"levelRequirement,omitempty"`
	Shop             bool          `json:"shop" firestore:"shop"`
	Description      *string       `json:"description,omitempty" firestore:"description,omitempty"`
	Cost             *int32        `json:"cost" firestore:"cost"`
	Stats            *StatModifier `json:"stats" firestore:"stats"`
	Boss             *string       `json:"boss" firestore:"boss"`
	SetBonusId       *string       `json:"setBonusId" firestore:"setBonusId"`
	RequiredClasses  *[]*string    `json:"requiredClasses" firestore:"requiredClasses"`
}

type InventoryItem struct {
	Name  string
	Count int
}

type ItemType struct {
	Type       string  `json:"itemType" firestore:"itemType"`
	WeaponType *string `json:"weaponType,omitempty" firestore:"weaponType,omitempty"`
}
