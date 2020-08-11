package models

type Monster struct {
	Name  string       `json:"name" firestore:"name"`
	Level int32        `json:"level" firestore:"level"`
	Exp   int32        `json:"exp" firestore:"exp"`
	Stats StatModifier `json:"stats" firestore:"stats"`
}
