package resolvers

import (
	"context"
	"lataleBotService/models"
)

type inventoryResponseResolver struct {
	inventory *models.Inventory
	message   *string
}

func (i *inventoryResponseResolver) Inventory(_ context.Context) *inventoryResolver {
	if i.inventory == nil {
		return nil
	}
	return &inventoryResolver{inventory: i.inventory}
}

func (i *inventoryResponseResolver) Message(_ context.Context) *string {
	return i.message
}
