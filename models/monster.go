package models

type Monster struct {
	Name  string       `json:"name" firestore:"name"`
	Level int32        `json:"level" firestore:"level"`
	Ely   int32        `json:"ely" firestore:"ely"`
	Rank  int32        `json:"rank" firestore:"rank"`
	Stats StatModifier `json:"stats" firestore:"stats"`
}
