package resolvers

import (
	"context"
	"lataleBotService/models"
)

type equipmentResolver struct {
	equipment models.Equipment
}

func (e *equipmentResolver) Weapon(_ context.Context) *string {
	return e.equipment.Weapon
}

func (e *equipmentResolver) Body(_ context.Context) *string {
	return e.equipment.Body
}

func (e *equipmentResolver) Glove(_ context.Context) *string {
	return e.equipment.Glove
}

func (e *equipmentResolver) Shoes(_ context.Context) *string {
	return e.equipment.Shoes
}
