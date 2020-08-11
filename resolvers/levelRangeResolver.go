package resolvers

import (
	"context"
	"lataleBotService/models"
)

type levelRangeResolver struct {
	levelRange models.LevelRange
}

func (l *levelRangeResolver) Max(_ context.Context) int32 {
	return l.levelRange.Max
}

func (l *levelRangeResolver) Min(_ context.Context) int32 {
	return l.levelRange.Min
}
