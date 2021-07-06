package models

import "time"

type User struct {
	Name               string                `json:"name" firestore:"name"`
	ID                 string                `json:"id" firestore:"id"`
	Ely                *int64                `json:"ely" firestore:"ely"`
	AscensionLevel     int32                 `json:"ascensionLevel" firestore:"ascensionLevel"`
	AscensionExp       int64                 `json:"ascensionExp" firestore:"ascensionExp"`
	LastActionTime     time.Time             `json:"lastActionTime" firestore:"lastActionTime"`
	LastBossActionTime time.Time             `json:"lastBossActionTime" firestore:"lastBossActionTime"`
	CurrentClass       string                `json:"currentClass" firestore:"currentClass"`
	ClassMap           map[string]*ClassInfo `json:"classInfo" firestore:"classInfo"`
	Inventory          Inventory             `json:"inventory,omitempty" firestore:"inventory,omitempty"`
	Party              *string               `json:"party,omitempty" firestore:"party,omitempty"`
	Classes            *[]*ClassInfo         `firestore:"classes,omitempty"`
	PartyMembers       *[]string
}

type Inventory struct {
	Equipment map[string]int `json:"equipment" firestore:"equipment"`
	Consume   map[string]int `json:"consume" firestore:"consume"`
	Event     map[string]int `json:"event" firestore:"event"`
}

type UserBlob struct {
	JobClass    *JobClass
	BaseStats   *StatModifier
	BattleStats *StatModifier
	User        *User
	CurrentHP   int
	MaxHP       int
	UserLevel   int32
	Weapon      string
	Summons     []Summons
	Buffs       map[string]Buff
	Debuffs     map[string]CrowdControlTrait
	HitCount    int
}

type MonsterBlob struct {
	Monster      Monster
	CurrentHP    int32
	Name         string
	Ely          float64
	Exp          float64
	StatModifier *StatModifier
	BattleStats  *StatModifier
	Rank         int32
	Buffs        map[string]Buff
	Debuffs      map[string]CrowdControlTrait
	Bound        bool
}
