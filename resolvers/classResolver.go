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

func (c *classResolver) Exp(_ context.Context) float64 {
	return c.classInfo.Exp
}

func (c *classResolver) Equipment(_ context.Context) *equipmentResolver {
	return &equipmentResolver{equipment: &c.classInfo.Equipment}
}
