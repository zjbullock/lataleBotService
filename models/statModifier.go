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
	TargetDefenseDecrease  float64 `json:"targetDefenseDecrease" firestore:"targetDefenseDecrease"`
}

func (s *StatModifier) AddBuffStats(buffModifiers StatModifier) StatModifier {
	return StatModifier{
		MaxDPS:                 buffModifiers.MaxDPS * s.MaxDPS,
		MinDPS:                 buffModifiers.MinDPS * s.MinDPS,
		Defense:                buffModifiers.Defense * s.Defense,
		HP:                     buffModifiers.HP * s.HP,
		CriticalDamageModifier: buffModifiers.CriticalDamageModifier * s.CriticalDamageModifier,
		CriticalRate:           buffModifiers.CriticalRate,
		Accuracy:               buffModifiers.Accuracy,
		Evasion:                buffModifiers.Evasion,
		SkillDamageModifier:    buffModifiers.SkillDamageModifier,
		SkillProcRate:          buffModifiers.SkillProcRate,
		Recovery:               buffModifiers.Recovery,
		TargetDefenseDecrease:  buffModifiers.TargetDefenseDecrease,
	}
}

func (s *StatModifier) AddStatModifier(stat StatModifier) {
	s.CriticalRate += stat.CriticalRate
	s.MaxDPS += stat.MaxDPS
	s.MinDPS += stat.MinDPS
	s.Defense += stat.Defense
	s.HP += stat.HP
	s.CriticalDamageModifier += stat.CriticalDamageModifier
	s.Accuracy += stat.Accuracy
	s.Evasion += stat.Evasion
	s.SkillProcRate += stat.SkillProcRate
	s.Recovery += stat.Recovery
	s.SkillDamageModifier += stat.SkillDamageModifier
	s.TargetDefenseDecrease += stat.TargetDefenseDecrease
}

func (s *StatModifier) AmplifyStatModifier(stat StatModifier) {
	s.CriticalRate *= stat.CriticalRate
	s.MaxDPS *= stat.MaxDPS
	s.MinDPS *= stat.MinDPS
	s.Defense *= stat.Defense
	s.HP *= stat.HP
	s.CriticalDamageModifier *= stat.CriticalDamageModifier
	s.Accuracy *= stat.Accuracy
	s.Evasion *= stat.Evasion
	s.SkillProcRate *= stat.SkillProcRate
	s.Recovery *= stat.Recovery
	s.SkillDamageModifier *= stat.SkillDamageModifier
	s.TargetDefenseDecrease *= stat.TargetDefenseDecrease
}

func (s *StatModifier) SubtractStatModifier(stat StatModifier) {
	s.CriticalRate -= stat.CriticalRate
	s.MaxDPS -= stat.MaxDPS
	s.MinDPS -= stat.MinDPS
	s.CriticalDamageModifier -= stat.CriticalDamageModifier
	s.Defense -= stat.Defense
	s.Accuracy -= stat.Accuracy
	s.Evasion -= stat.Evasion
	s.HP -= stat.HP
	s.SkillProcRate -= stat.SkillProcRate
	s.Recovery -= stat.Recovery
	s.SkillDamageModifier -= stat.SkillDamageModifier
	s.TargetDefenseDecrease -= stat.TargetDefenseDecrease
}
