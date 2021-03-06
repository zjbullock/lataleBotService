package resolvers

import (
	"context"
	"lataleBotService/models"
)

type itemInfoResolver struct {
	item models.Item
}

func (i *itemInfoResolver) Name(_ context.Context) string {
	return i.item.Name
}

func (i *itemInfoResolver) Type(_ context.Context) *itemTypeResolver {
	return &itemTypeResolver{itemType: &i.item.Type}
}

func (i *itemInfoResolver) LevelRequirement(_ context.Context) *float64 {
	return i.item.LevelRequirement
}

func (i *itemInfoResolver) AscensionLevelRequirement(_ context.Context) *float64 {
	return i.item.AscensionLevelRequirement
}

func (i *itemInfoResolver) Shop(_ context.Context) bool {
	return i.item.Shop
}

func (i *itemInfoResolver) Description(_ context.Context) *string {
	return i.item.Description
}

func (i *itemInfoResolver) Cost(_ context.Context) *int32 {
	return i.item.Cost
}

func (i *itemInfoResolver) Boss(_ context.Context) *string {
	return i.item.Boss
}

func (i *itemInfoResolver) Stats(_ context.Context) *statResolver {
	if i.item.Stats == nil {
		return nil
	}
	return &statResolver{stat: i.item.Stats}
}

func (i *itemInfoResolver) SetBonusId(_ context.Context) *string {
	return i.item.SetBonusId
}

func (i *itemInfoResolver) RequiredClasses(_ context.Context) *[]*string {
	return i.item.RequiredClasses
}
