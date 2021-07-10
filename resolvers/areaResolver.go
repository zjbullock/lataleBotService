package resolvers

import (
	"context"
	"lataleBotService/models"
)

type areaResolver struct {
	area models.Area
}

func (a *areaResolver) Name(_ context.Context) string {
	return a.area.Name
}

func (a *areaResolver) ID(_ context.Context) string {
	return a.area.ID
}

func (a *areaResolver) LevelRange(_ context.Context) *levelRangeResolver {
	return &levelRangeResolver{levelRange: a.area.LevelRange}
}

func (a *areaResolver) AscensionRange(_ context.Context) *levelRangeResolver {
	if a.area.AscensionRange != nil {
		return &levelRangeResolver{levelRange: *a.area.AscensionRange}
	}
	return nil
}

func (a *areaResolver) Monsters(_ context.Context) []*monsterResolver {
	var monsters []*monsterResolver
	for _, monster := range a.area.Monsters {
		monsters = append(monsters, &monsterResolver{monsterInfo: monster})
	}
	return monsters
}
