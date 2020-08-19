package models

import "time"

type User struct {
	Name           string               `json:"name" firestore:"name"`
	ID             string               `json:"id" firestore:"id"`
	Ely            *int32               `json:"ely" firestore:"ely"`
	LastActionTime time.Time            `json:"lastActionTime" firestore:"lastActionTime"`
	CurrentClass   string               `json:"currentClass" firestore:"currentClass"`
	ClassMap       map[string]ClassInfo `json:"classInfo" firestore:"classInfo"`
	Classes        *[]*ClassInfo        `firestore:"classes,omitempty"`
}
