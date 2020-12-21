package resolvers

import (
	"context"
	"lataleBotService/models"
	"sort"
)

type inventoryResolver struct {
	inventory *models.Inventory
}

type inventoryItemResolver struct {
	basicItem *models.InventoryItem
}

func (i *inventoryResolver) Equipment(_ context.Context) *[]*inventoryItemResolver {
	var equips []*inventoryItemResolver
	for name, count := range i.inventory.Equipment {
		equips = append(equips, &inventoryItemResolver{basicItem: &models.InventoryItem{Name: name, Count: count}})
	}
	if len(equips) == 0 {
		return nil
	}
	sort.Slice(equips, func(i, j int) bool {
		return equips[j].basicItem.Name > equips[i].basicItem.Name
	})
	return &equips
}

func (i *inventoryResolver) Consume(_ context.Context) *[]*inventoryItemResolver {
	var consumables []*inventoryItemResolver
	for name, count := range i.inventory.Consume {
		consumables = append(consumables, &inventoryItemResolver{basicItem: &models.InventoryItem{Name: name, Count: count}})
	}
	if len(consumables) == 0 {
		return nil
	}
	return &consumables
}

func (i *inventoryResolver) Event(_ context.Context) *[]*inventoryItemResolver {
	var eventItems []*inventoryItemResolver
	for name, count := range i.inventory.Event {
		eventItems = append(eventItems, &inventoryItemResolver{basicItem: &models.InventoryItem{Name: name, Count: count}})
	}
	if len(eventItems) == 0 {
		return nil
	}
	return &eventItems
}

func (i *inventoryItemResolver) Name(_ context.Context) string {
	return i.basicItem.Name
}

func (i *inventoryItemResolver) Count(_ context.Context) int32 {
	return int32(i.basicItem.Count)
}
