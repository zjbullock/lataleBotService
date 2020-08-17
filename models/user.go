package models

type User struct {
	Name         string               `json:"name" firestore:"name"`
	ID           string               `json:"id" firestore:"id"`
	Ely          *int32               `json:"ely" firestore:"ely"`
	CurrentClass string               `json:"currentClass" firestore:"currentClass"`
	ClassMap     map[string]ClassInfo `json:"classInfo" firestore:"classInfo"`
	Classes      *[]*ClassInfo        `firestore:"classes,omitempty"`
}
