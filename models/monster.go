package models

type Monster struct {
	Name           string        `json:"name" firestore:"name"`
	BossTitle      *string       `json:"bossTitle" firestore:"bossTitle"`
	Level          int32         `json:"level" firestore:"level"`
	AscensionLevel *int32        `json:"ascensionLevel" firestore:"ascensionLevel"`
	Ely            float64       `json:"ely" firestore:"ely"`
	Exp            float64       `json:"exp" firestore:"exp"`
	Rank           int32         `json:"rank" firestore:"rank"`
	Stats          StatModifier  `json:"stats" firestore:"stats"`
	BossBonus      *BossBonus    `json:"bossBonus" firestore:"bossBonus"`
	Skills         *[]*BossSkill `json:"skills,omitempty" firestore:"skills,omitempty"`
	IdleTime       *float64      `json:"idleTime" firestore:"idleTime,omitempty"`
	Phases         *[]string     `json:"phases" firestore:"phases,omitempty"`
	IdlePhrase     *string       `json:"idlePhrase" firestore:"idlePhrase,omitempty"`
	DropRange      *LevelRange   `json:"dropRange,omitempty" firestore:"dropRange,omitempty"`
}

type BossSkill struct {
	Name                string        `json:"name" firestore:"name"`
	Quote               string        `json:"quote" firestore:"quote"`
	CrowdControl        *int32        `json:"crowdControl" firestore:"crowdControl"`
	CrowdControlStatus  *string       `json:"crowdControlStatus" firestore:"crowdControlStatus"`
	Debuff              *StatModifier `json:"debuff,omitempty" firestore:"debuff,omitempty"`
	AoE                 bool          `json:"aoe" firestore:"aoe"`
	Bind                bool          `json:"bind" firestore:"bind"`
	SkillDamageModifier float64       `json:"skillDamageModifier" firestore:"skillDamageModifier"`
	CoolDown            int32         `json:"coolDown" firestore:"coolDown"`
	CurrentCoolDown     *int32        `json:"currentCoolDown" firestore:"currentCoolDown,omitempty"`
}
