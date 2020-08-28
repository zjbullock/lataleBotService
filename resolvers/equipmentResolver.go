package resolvers

import (
	"context"
	"lataleBotService/models"
)

type equipmentResolver struct {
	equipment *models.Equipment
}

func (e *equipmentResolver) Body(_ context.Context) string {
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

func (e *equipmentResolver) Bindi(_ context.Context) string {
	return e.equipment.EquipmentNames[4]
}

func (e *equipmentResolver) Glasses(_ context.Context) string {
	return e.equipment.EquipmentNames[5]
}

func (e *equipmentResolver) Earrings(_ context.Context) string {
	return e.equipment.EquipmentNames[6]
}

func (e *equipmentResolver) Ring(_ context.Context) string {
	return e.equipment.EquipmentNames[7]
}

func (e *equipmentResolver) Cloak(_ context.Context) string {
	return e.equipment.EquipmentNames[8]
}

func (e *equipmentResolver) Stockings(_ context.Context) string {
	return e.equipment.EquipmentNames[9]
}
