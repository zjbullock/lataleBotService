package resolvers

import (
	"context"
	"lataleBotService/models"
)

type newUserResolver struct {
	newUserResponse models.NewUserResponse
}

func (n *newUserResolver) ID(_ context.Context) *string {
	return n.newUserResponse.ID
}

func (n *newUserResolver) Message(_ context.Context) *string {
	return n.newUserResponse.Message
}
