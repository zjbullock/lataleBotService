package resolvers

import (
	"context"
	"lataleBotService/models"
)

type setBonusResolver struct {
	setBonus models.SetBonus
}

func (s *setBonusResolver) Name(_ context.Context) *string {
	return &s.setBonus.Name
}

func (s *setBonusResolver) Id(_ context.Context) *string {
	return &s.setBonus.Id
}

func (s *setBonusResolver) RequiredPieces(_ context.Context) *int32 {
	return &s.setBonus.RequiredPieces
}

func (s *setBonusResolver) CurrentlyEquipped(_ context.Context) *int32 {
	return &s.setBonus.CurrentlyEquipped
}

func (s *setBonusResolver) Bonus(_ context.Context) *statResolver {
	return &statResolver{stat: s.setBonus.Bonus}
}
