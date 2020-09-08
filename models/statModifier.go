package models

type StatModifier struct {
	CriticalRate           float64 `json:"criticalRate" firestore:"criticalRate"`
	MaxDPS                 float64 `json:"maxDps" firestore:"maxDps"`
	MinDPS                 float64 `json:"minDps" firestore:"minDps"`
	CriticalDamageModifier float64 `json:"criticalDamageModifier" firestore:"criticalDamageModifier"`
	Defense                float64 `json:"defense" firestore:"defense"`
	Accuracy               float64 `json:"accuracy" firestore:"accuracy"`
	Evasion                float64 `json:"evasion" firestore:"evasion"`
	HP                     float64 `json:"hp" firestore:"hp"`
	SkillProcRate          float64 `json:"skillProcRate" firestore:"skillProcRate"`
	Recovery               float64 `json:"recovery" firestore:"recovery"`
	SkillDamageModifier    float64 `json:"skillDamageModifier" firestore:"skillDamageModifier"`
}

func (s StatModifier) AddStatModifier(stat StatModifier) StatModifier {
	newStats := s
	newStats.CriticalRate += stat.CriticalRate
	newStats.MaxDPS += stat.MaxDPS
	newStats.MinDPS += stat.MinDPS
	newStats.CriticalDamageModifier += stat.CriticalDamageModifier
	newStats.Defense += stat.Defense
	newStats.Accuracy += stat.Accuracy
	newStats.Evasion += stat.Evasion
	newStats.HP += stat.HP
	newStats.SkillProcRate += stat.SkillProcRate
	newStats.Recovery += stat.Recovery
	newStats.SkillDamageModifier += stat.SkillDamageModifier
	return newStats
}
