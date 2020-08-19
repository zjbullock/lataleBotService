package models

type ClassInfo struct {
	Name          string    `json:"name" firestore:"name"`
	Level         int32     `json:"level" firestore:"level"`
	Exp           float64   `json:"exp" firestore:"exp"`
	CurrentWeapon string    `json:"currentWeapon" firestore:"currentWeapon"`
	Equipment     Equipment `json:"equipment" firestore:"equipment"`
}
