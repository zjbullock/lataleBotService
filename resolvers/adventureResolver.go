package resolvers

import "context"

type adventureResponseResolver struct {
	log     *[]string
	message *string
}

func (a *adventureResponseResolver) Log(_ context.Context) *[]string {
	return a.log
}

func (a *adventureResponseResolver) Message(_ context.Context) *string {
	return a.message
}
