package models

type Area struct {
	Name       string     `json:"name" firestore:"name"`
	LevelRange LevelRange `json:"levelRange": firestore:"levelRange"`
	Monsters   []Monster  `json:"monsters" firestore:"monsters"`
}

type LevelRange struct {
	Max int32 `json:"max" firestore:"max"`
	Min int32 `json:"min" firestore:"min"`
}
