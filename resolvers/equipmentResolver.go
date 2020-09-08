package resolvers

import (
	"context"
	"lataleBotService/models"
)

type equipmentResolver struct {
	equipment *models.Equipment
}

func (e *equipmentResolver) Top(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Top}
}

func (e *equipmentResolver) Bottom(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Bottom}
}

func (e *equipmentResolver) Headpiece(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Headpiece}
}

func (e *equipmentResolver) Gloves(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Glove}
}

func (e *equipmentResolver) Boots(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Shoes}
}

func (e *equipmentResolver) Weapon(_ context.Context) *itemInfoResolver {
	return &itemInfoResolver{item: e.equipment.Weapon}
}

func (e *equipmentResolver) Bindi(_ context.Context) *itemInfoResolver {
	if e.equipment.Bindi == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Bindi}
}

func (e *equipmentResolver) Glasses(_ context.Context) *itemInfoResolver {
	if e.equipment.Glasses == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Glasses}
}

func (e *equipmentResolver) Earrings(_ context.Context) *itemInfoResolver {
	if e.equipment.Earring == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Earring}
}

func (e *equipmentResolver) Ring(_ context.Context) *itemInfoResolver {
	if e.equipment.Ring == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Ring}
}

func (e *equipmentResolver) Cloak(_ context.Context) *itemInfoResolver {
	if e.equipment.Cloak == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Cloak}
}

func (e *equipmentResolver) Stockings(_ context.Context) *itemInfoResolver {
	if e.equipment.Stockings == nil {
		return nil
	}
	return &itemInfoResolver{item: *e.equipment.Stockings}
}
