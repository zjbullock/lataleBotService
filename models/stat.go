package models

type Stat struct {
	CriticalRate           float64 `json:"criticalRate" firestore:"criticalRate"`
	DPS                    float64 `json:"dps" firestore:"dps"`
	CriticalDamageModifier float64 `json:"criticalDamageModifier" firestore:"criticalDamageModifier"`
	Defense                float64 `json:"defense" firestore:"defense"`
	Accuracy               float64 `json:"accuracy" firestore:"accuracy"`
	Evasion                float64 `json:"evasion" firestore:"evasion"`
	HP                     float64 `json:"hp" firestore:"hp"`
	SkillProcRate          float64 `json:"skillProcRate" firestore:"skillProcRate"`
	Recovery               float64 `json:"recovery" firestore:"recovery"`
}
