package models

import "time"

type User struct {
	Name               string                `json:"name" firestore:"name"`
	ID                 string                `json:"id" firestore:"id"`
	Ely                *int32                `json:"ely" firestore:"ely"`
	LastActionTime     time.Time             `json:"lastActionTime" firestore:"lastActionTime"`
	LastBossActionTime time.Time             `json:"lastBossActionTime" firestore:"lastBossActionTime"`
	CurrentClass       string                `json:"currentClass" firestore:"currentClass"`
	ClassMap           map[string]*ClassInfo `json:"classInfo" firestore:"classInfo"`
	Classes            *[]*ClassInfo         `firestore:"classes,omitempty"`
	Inventory          Inventory             `json:"inventory" firestore:"inventory"`
	Party              *string               `json:"party,omitempty" firestore:"party,omitempty"`
	PartyMembers       *[]string
}

type Inventory struct {
	Equipment map[string]interface{} `json:"equipment" firestore:"equipment"`
	Consume   map[string]interface{} `json:"consume" firestore:"consume"`
	Event     map[string]interface{} `json:"event" firestore:"event"`
}

type UserBlob struct {
	JobClass           *JobClass
	StatModifier       *StatModifier
	User               *User
	CurrentHP          int
	MaxHP              int
	UserLevel          int32
	Weapon             string
	CrowdControlled    *int32
	CrowdControlStatus *string
}

type MonsterBlob struct {
	CurrentHP    int32
	Name         string
	Ely          int32
	Exp          int32
	StatModifier *StatModifier
}
