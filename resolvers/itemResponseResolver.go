package resolvers

import (
	"context"
	"lataleBotService/models"
)

type itemResponseResolver struct {
	item    *models.Item
	message *string
}

func (i *itemResponseResolver) Item(_ context.Context) *itemInfoResolver {
	if i.item == nil {
		return nil
	}
	return &itemInfoResolver{item: *i.item}
}

func (i *itemResponseResolver) Message(_ context.Context) *string {
	return i.message
}
