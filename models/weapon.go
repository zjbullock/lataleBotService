package models

type Weapon struct {
	Name   string   `json:"name" firestore:"name"`
	Skills []string `json:"skills" firestore:"skills"`
}
