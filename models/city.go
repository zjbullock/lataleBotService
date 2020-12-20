package models

type City struct {
	ID         int        `json:"id" firestore:"id"`
	Name       string     `json:"name" firestore:"name"`
	LevelRange LevelRange `json:"levelRange" firestore:"levelRange"`
}
