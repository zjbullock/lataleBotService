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

func (c *classResolver) Level(_ context.Context) float64 {
	return float64(c.classInfo.Level)
}

func (c *classResolver) Exp(_ context.Context) float64 {
	return float64(c.classInfo.Exp)
}

func (c *classResolver) Equipment(_ context.Context) *equipmentResolver {
	return &equipmentResolver{equipment: &c.classInfo.Equipment}
}

func (c *classResolver) SetBonuses(_ context.Context) *[]string {
	var setBonuses []string
	if c.classInfo.SetBonuses != nil && len(c.classInfo.SetBonuses) > 0 {
		var setBonusSorting []struct {
			ID   string
			Name string
		}
		for _, bonus := range c.classInfo.SetBonuses {
			setBonusSorting = append(setBonusSorting, struct {
				ID   string
				Name string
			}{ID: bonus.Id, Name: bonus.Name})
		}
		sort.Slice(setBonusSorting, func(i, j int) bool {
			return setBonusSorting[i].ID < setBonusSorting[j].ID
		})
		for _, bonus := range setBonusSorting {
			setBonuses = append(setBonuses, fmt.Sprintf("%v		|		%s", bonus.ID, bonus.Name))
		}
	} else {
		return nil
	}
	return &setBonuses
}

func (c *classResolver) BossBonuses(_ context.Context) *[]string {
	var bossBonuses []string
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
