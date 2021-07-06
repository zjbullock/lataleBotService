package models

type ClassInfo struct {
	Name              string                `json:"name" firestore:"name"`
	Level             int32                 `json:"level" firestore:"level"`
	Ascension         int32                 `json:"ascension" firestore:"ascension"`
	Exp               int64                 `json:"exp" firestore:"exp"`
	CurrentWeapon     *string               `json:"currentWeapon,omitempty" firestore:"currentWeapon,omitempty"`
	OldEquipmentSheet *OldEquipmentSystem   `json:"equipment,omitempty" firestore:"equipment,omitempty"`
	Equipment         Equipment             `json:"currentEquips" firestore:"currentEquips"`
	SetBonuses        map[string]*SetBonus  `json:"setBonuses,omitempty" firestore:"setBonuses,omitempty"`
	BossBonuses       map[string]*BossBonus `json:"bossBonuses" firestore:"bossBonuses"`
}
