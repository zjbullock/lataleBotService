package resolvers

import (
	"context"
	"fmt"
	"lataleBotService/models"
)

type equipmentResolver struct {
	equipment *models.Equipment
}

func (e *equipmentResolver) Body(_ context.Context) string {
	fmt.Println(e.equipment.EquipmentNames)
	return e.equipment.EquipmentNames[0]
}

func (e *equipmentResolver) Glove(_ context.Context) string {
	return e.equipment.EquipmentNames[1]
}

func (e *equipmentResolver) Shoes(_ context.Context) string {
	return e.equipment.EquipmentNames[2]
}

func (e *equipmentResolver) Weapon(_ context.Context) string {
	return e.equipment.EquipmentNames[3]
}
