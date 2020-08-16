package resolvers

import (
	"context"
	"lataleBotService/models"
)

type statResponseResolver struct {
	stat    *models.StatModifier
	message *string
}

func (s *statResponseResolver) Stat(_ context.Context) *statResolver {
	if s.stat == nil {
		return nil
	}
	return &statResolver{stat: s.stat}
}

func (s *statResponseResolver) Message(_ context.Context) *string {
	return s.message
}
