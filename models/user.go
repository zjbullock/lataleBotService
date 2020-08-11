package models

type User struct {
	Name         string       `json:"name" firestore:"name"`
	Ely          *int32       `json:"ely" firestore:"ely"`
	CurrentClass string       `json:"currentClass" firestore:"currentClass"`
	CurrentLevel *int32       `json:"currentLevel" firestore:"currentLevel"`
	Classes      []*ClassInfo `json:"classInfo" firestore:"classInfo"`
}
