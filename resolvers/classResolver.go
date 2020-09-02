package resolvers

import (
	"context"
	"fmt"
	"lataleBotService/models"
	"sort"
)

type classResolver struct {
	classInfo models.ClassInfo
}

func (c *classResolver) Name(_ context.Context) string {
	return c.classInfo.Name
}

func (c *classResolver) Level(_ context.Context) int32 {
	return c.classInfo.Level
}

func (c *classResolver) Exp(_ context.Context) int32 {
	return c.classInfo.Exp
}

func (c *classResolver) Equipment(_ context.Context) *equipmentResolver {
	return &equipmentResolver{equipment: &c.classInfo.Equipment}
}

func (c *classResolver) BossBonuses(_ context.Context) *[]string {
	var bossBonuses []string
	bossBonuses = append(bossBonuses, fmt.Sprintf("ID:		|		Name:"))
	if c.classInfo.BossBonuses != nil {
		var bossBonusSorting []struct {
			ID   int32
			Name string
		}
		for _, bonus := range c.classInfo.BossBonuses {
			bossBonusSorting = append(bossBonusSorting, struct {
				ID   int32
				Name string
			}{ID: bonus.ID, Name: bonus.Name})
		}
		sort.Slice(bossBonusSorting, func(i, j int) bool {
			return bossBonusSorting[i].ID < bossBonusSorting[j].ID
		})
		for _, bonus := range bossBonusSorting {
			bossBonuses = append(bossBonuses, fmt.Sprintf("%v		|		%s", bonus.ID, bonus.Name))
		}
	} else {
		return nil
	}
	return &bossBonuses
}
