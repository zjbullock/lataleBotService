package models

type Area struct {
	Name       string     `json:"name" firestore:"name"`
	ID         string     `json:"id" firestore:"id"`
	LevelRange LevelRange `json:"levelRange" firestore:"levelRange"`
	Monsters   []Monster  `json:"monsters" firestore:"monsters"`
	DropRange  LevelRange `json:"dropRange" firestore:"dropRange"`
}

type LevelRange struct {
	Max int32 `json:"max" firestore:"max"`
	Min int32 `json:"min" firestore:"min"`
}
