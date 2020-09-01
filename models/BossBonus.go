package models

type BossBonus struct {
	Name                   string   `json:"name" firestore:"name"`
	ID                     int32    `json:"id" firestore:"id"`
	BossDropChance         *float64 `json:"bossDropChance" firestore:"bossDropChance"`
	CriticalRate           float64  `json:"criticalRate" firestore:"criticalRate"`
	MaxDPS                 float64  `json:"maxDps" firestore:"maxDps"`
	MinDPS                 float64  `json:"minDps" firestore:"minDps"`
	CriticalDamageModifier float64  `json:"criticalDamageModifier" firestore:"criticalDamageModifier"`
	Defense                float64  `json:"defense" firestore:"defense"`
	Accuracy               float64  `json:"accuracy" firestore:"accuracy"`
	Evasion                float64  `json:"evasion" firestore:"evasion"`
	HP                     float64  `json:"hp" firestore:"hp"`
	SkillProcRate          float64  `json:"skillProcRate" firestore:"skillProcRate"`
	SkillDamageModifier    float64  `json:"skillDamageModifier" firestore:"skillDamageModifier"`
	Recovery               float64  `json:"recovery" firestore:"recovery"`
}
