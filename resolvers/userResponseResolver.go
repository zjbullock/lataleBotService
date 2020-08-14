package resolvers

import (
	"context"
	"lataleBotService/models"
)

type userResponseResolver struct {
	user    *models.User
	message *string
}

func (u *userResponseResolver) User(_ context.Context) *userResolver {
	return &userResolver{user: u.user}
}

func (u *userResponseResolver) Message(_ context.Context) *string {
	return u.message
}
