package models

type ClassInfo struct {
	Name          string                `json:"name" firestore:"name"`
	Level         int64                 `json:"level" firestore:"level"`
	Exp           int64                 `json:"exp" firestore:"exp"`
	CurrentWeapon string                `json:"currentWeapon" firestore:"currentWeapon"`
	Equipment     Equipment             `json:"equipment" firestore:"equipment"`
	BossBonuses   map[string]*BossBonus `json:"bossBonuses" firestore:"bossBonuses"`
}
