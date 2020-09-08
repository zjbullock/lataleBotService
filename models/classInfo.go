package models

type ClassInfo struct {
	Name              string                `json:"name" firestore:"name"`
	Level             int64                 `json:"level" firestore:"level"`
	Exp               int64                 `json:"exp" firestore:"exp"`
	CurrentWeapon     *string               `json:"currentWeapon,omitempty" firestore:"currentWeapon,omitempty"`
	OldEquipmentSheet *OldEquipmentSystem   `json:"equipment,omitempty" firestore:"equipment,omitempty"`
	Equipment         Equipment             `json:"currentEquips" firestore:"currentEquips"`
	BossBonuses       map[string]*BossBonus `json:"bossBonuses" firestore:"bossBonuses"`
}
