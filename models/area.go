package models

type Area struct {
	Name       string     `json:"name" firestore:"name"`
	ID         string     `json:"id" firestore:"id"`
	LevelRange LevelRange `json:"levelRange" firestore:"levelRange"`
	Monsters   []Monster  `json:"monsters" firestore:"monsters"`
}

type LevelRange struct {
	Max int64 `json:"max" firestore:"max"`
	Min int64 `json:"min" firestore:"min"`
}
