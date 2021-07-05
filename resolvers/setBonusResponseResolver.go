package resolvers

import (
	"context"
	"lataleBotService/models"
)

type setBonusResponseResolver struct {
	setBonus *models.SetBonus
	message  *string
}

func (s *setBonusResponseResolver) SetBonus(_ context.Context) *setBonusResolver {
	if s.setBonus == nil {
		return nil
	}
	return &setBonusResolver{setBonus: *s.setBonus}
}

func (s *setBonusResponseResolver) Message(_ context.Context) *string {
	return s.message
}
