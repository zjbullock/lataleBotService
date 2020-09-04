package resolvers

import (
	"context"
	"lataleBotService/models"
)

type levelRangeResolver struct {
	levelRange models.LevelRange
}

func (l *levelRangeResolver) Max(_ context.Context) float64 {
	return float64(l.levelRange.Max)
}

func (l *levelRangeResolver) Min(_ context.Context) float64 {
	return float64(l.levelRange.Min)
}
