package resolvers

import (
	"context"
	"lataleBotService/models"
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
	if c.classInfo.BossBonuses != nil {
		for _, bonus := range c.classInfo.BossBonuses {
			bossBonuses = append(bossBonuses, bonus.Name)
		}
	} else {
		return nil
	}
	return &bossBonuses
}
