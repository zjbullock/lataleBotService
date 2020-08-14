package models

type User struct {
	Name          string               `json:"name" firestore:"name"`
	Ely           *int32               `json:"ely" firestore:"ely"`
	CurrentClass  string               `json:"currentClass" firestore:"currentClass"`
	CurrentLevel  *int32               `json:"currentLevel" firestore:"currentLevel"`
	CurrentWeapon string               `json:"currentWeapon" firestore:"currentWeapon"`
	ClassMap      map[string]ClassInfo `json:"classInfo" firestore:"classInfo"`
	Classes       *[]*ClassInfo        `firestore:"classes,omitempty"`
}
