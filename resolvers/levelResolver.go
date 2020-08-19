package resolvers

import (
	"context"
	"lataleBotService/models"
)

type levelResolver struct {
	level models.Level
}

func (l *levelResolver) Value(_ context.Context) int32 {
	return l.level.Value
}

func (l *levelResolver) Exp(_ context.Context) float64 {
	return l.level.Exp
}
