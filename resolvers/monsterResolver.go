package resolvers

import (
	"context"
	"lataleBotService/models"
)

type monsterResolver struct {
	monsterInfo models.Monster
}

func (m *monsterResolver) Name(_ context.Context) string {
	return m.monsterInfo.Name
}

func (m *monsterResolver) Level(_ context.Context) int32 {
	return m.monsterInfo.Level
}

func (m *monsterResolver) Ely(_ context.Context) int32 {
	return m.monsterInfo.Ely
}

func (m *monsterResolver) Rank(_ context.Context) int32 {
	return m.monsterInfo.Rank
}

func (m *monsterResolver) Stats(_ context.Context) *statResolver {
	return &statResolver{stat: &m.monsterInfo.Stats}
}
