package models

type Monster struct {
	Name  string       `json:"name" firestore:"name"`
	Level int32        `json:"level" firestore:"level"`
	Exp   int32        `json:"exp" firestore:"exp"`
	Ely   int32        `json:"ely" firestore:"ely"`
	Stats StatModifier `json:"stats" firestore:"stats"`
}
