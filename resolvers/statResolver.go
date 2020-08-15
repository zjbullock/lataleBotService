package resolvers

import (
	"context"
	"lataleBotService/models"
	"math"
)

type statResolver struct {
	stat models.StatModifier
}

func (s *statResolver) CriticalRate(_ context.Context) float64 {
	return s.stat.CriticalRate
}

func (s *statResolver) MaxDPS(_ context.Context) int32 {
	return int32(math.Round(s.stat.MaxDPS))
}

func (s *statResolver) MinDPS(_ context.Context) int32 {
	return int32(math.Round(s.stat.MinDPS))
}

func (s *statResolver) CriticalDamageModifier(_ context.Context) float64 {
	return s.stat.CriticalDamageModifier
}

func (s *statResolver) Defense(_ context.Context) int32 {
	return int32(math.Round(s.stat.Defense))
}

func (s *statResolver) Accuracy(_ context.Context) float64 {
	return s.stat.Accuracy
}

func (s *statResolver) Evasion(_ context.Context) float64 {
	return s.stat.Evasion
}

func (s *statResolver) HP(_ context.Context) int32 {
	return int32(math.Round(s.stat.HP))
}

func (s *statResolver) SkillProcRate(_ context.Context) float64 {
	return s.stat.SkillProcRate
}

func (s *statResolver) Recovery(_ context.Context) float64 {
	return s.stat.Recovery
}
