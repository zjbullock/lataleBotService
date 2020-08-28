package models

import "time"

type User struct {
	Name           string                `json:"name" firestore:"name"`
	ID             string                `json:"id" firestore:"id"`
	Ely            *int32                `json:"ely" firestore:"ely"`
	LastActionTime time.Time             `json:"lastActionTime" firestore:"lastActionTime"`
	CurrentClass   string                `json:"currentClass" firestore:"currentClass"`
	ClassMap       map[string]*ClassInfo `json:"classInfo" firestore:"classInfo"`
	Classes        *[]*ClassInfo         `firestore:"classes,omitempty"`
	Party          *string               `json:"party,omitempty" firestore:"party,omitempty"`
	PartyMembers   *[]string
}

type UserBlob struct {
	JobClass     *JobClass
	StatModifier *StatModifier
	User         *User
	CurrentHP    int
	MaxHP        int
	UserLevel    int32
	Weapon       string
}

type MonsterBlob struct {
	CurrentHP    int32
	Name         string
	Ely          int32
	Exp          int32
	StatModifier *StatModifier
}
