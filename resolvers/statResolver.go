package resolvers

import (
	"context"
	"lataleBotService/models"
	"math"
)

type statResolver struct {
	stat *models.StatModifier
}

func (s *statResolver) CriticalRate(_ context.Context) *float64 {
	return &s.stat.CriticalRate
}

func (s *statResolver) MaxDPS(_ context.Context) *int32 {
	max := int32(math.Round(s.stat.MaxDPS))
	return &max
}

func (s *statResolver) MinDPS(_ context.Context) *int32 {
	min := int32(math.Round(s.stat.MinDPS))
	return &min
}

func (s *statResolver) CriticalDamageModifier(_ context.Context) *float64 {
	return &s.stat.CriticalDamageModifier
}

func (s *statResolver) Defense(_ context.Context) *int32 {
	def := int32(math.Round(s.stat.Defense))
	return &def
}

func (s *statResolver) Accuracy(_ context.Context) *float64 {
	return &s.stat.Accuracy
}

func (s *statResolver) Evasion(_ context.Context) *float64 {
	return &s.stat.Evasion
}

func (s *statResolver) HP(_ context.Context) *int32 {
	hp := int32(math.Round(s.stat.HP))
	return &hp
}

func (s *statResolver) SkillProcRate(_ context.Context) *float64 {
	return &s.stat.SkillProcRate
}

func (s *statResolver) Recovery(_ context.Context) *float64 {
	return &s.stat.Recovery
}
