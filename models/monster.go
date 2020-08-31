package models

type Monster struct {
	Name       string        `json:"name" firestore:"name"`
	Level      int32         `json:"level" firestore:"level"`
	Ely        int32         `json:"ely" firestore:"ely"`
	Exp        int32         `json:"exp" firestore:"exp"`
	Rank       int32         `json:"rank" firestore:"rank"`
	Stats      StatModifier  `json:"stats" firestore:"stats"`
	BossBonus  *int32        `json:"bossBonus" firestore:"bossBonus"`
	Skills     *[]*BossSkill `json:"skills,omitempty" firestore:"skills,omitempty"`
	IdleTime   *float64      `json:"idleTime" firestore:"idleTime"`
	Phases     *[]string     `json:"phases" firestore:"phases,omitempty"`
	IdlePhrase *string       `json:"idlePhrase" firestore:"idlePhrase,omitempty"`
}

type BossSkill struct {
	Name                string  `json:"name" firestore:"name"`
	Quote               string  `json:"quote" firestore:"quote"`
	CrowdControl        *int32  `json:"crowdControl" firestore:"crowdControl"`
	AoE                 bool    `json:"aoe" firestore:"aoe"`
	SkillDamageModifier float64 `json:"skillDamageModifier" firestore:"skillDamageModifier"`
	CoolDown            int32   `json:"coolDown" firestore:"coolDown"`
	CurrentCoolDown     *int32  `json:"currentCoolDown" firestore:"currentCoolDown,omitempty"`
}
