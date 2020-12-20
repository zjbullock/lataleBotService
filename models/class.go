package models

import "lataleBotService/globals"

type JobClass struct {
	Name             string       `json:"name" firestore:"name"`
	LevelRequirement int32        `json:"levelRequirement" firestore:"levelRequirement"`
	ClassRequirement *string      `json:"classRequirement" firestore:"classRequirement"`
	Tier             int32        `json:"tier" firestore:"tier"`
	Weapons          []Weapon     `json:"weapons" firestore:"weapons"`
	Stats            StatModifier `json:"stats" firestore:"stats"`
	Description      string       `json:"description" firestore:"description"`
	Trait            *Trait       `json:"trait,omitempty" firestore:"trait,omitempty"`
}

type Trait struct {
	Name           string             `json:"name" firestore:"name"`
	Description    string             `json:"description" firestore:"description,omitempty"`
	Type           globals.TraitType  `json:"type" firestore:"type"`
	HPTrigger      *float64           `json:"hpTrigger,omitempty" firestore:"hpTrigger,omitempty"`
	ActivationRate *float64           `json:"activationRate,omitempty" firestore:"activationRate,omitempty"`
	UsageCount     *int32             `json:"usageCount,omitempty" firestore:"usageCount,omitempty"`
	CrowdControl   *CrowdControlTrait `json:"crowdControl,omitempty" firestore:"crowdControl,omitempty"`
	Battle         *BattleTrait       `json:"battleTrait,omitempty" firestore:"battleTrait,omitempty"`
}

type CrowdControlTrait struct {
	Type             string `json:"type" firestore:"type"`
	CrowdControlTime int32  `json:"crowdControlTime,omitempty" firestore:"crowdControlTime"`
}

type BattleTrait struct {
	AoE        bool  `json:"aoe,omitempty" firestore:"aoe,omitempty"`
	HitCounter int32 `json:"hitCounter" firestore:"hitCounter"`
	Buff       Buff  `json:"buff,omitempty" firestore:"buff,omitempty"`
}

type Buff struct {
	StatModifier StatModifier `json:"stats" firestore:"stats"`
	Duration     *int32       `json:"duration,omitempty" firestore:"duration,omitempty"`
}
