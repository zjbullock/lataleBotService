package resolvers

import (
	"context"
	"lataleBotService/models"
)

type userResolver struct {
	user models.User
}

func (u *userResolver) Name(_ context.Context) string {
	return u.user.Name
}

func (u *userResolver) Ely(_ context.Context) *int32 {
	return u.user.Ely
}

func (u *userResolver) CurrentClass(_ context.Context) string {
	return u.user.CurrentClass
}

func (u *userResolver) CurrentLevel(_ context.Context) *int32 {
	return u.user.CurrentLevel
}

func (u *userResolver) Classes(_ context.Context) []*classResolver {
	var classes []*classResolver
	for _, class := range u.user.Classes {
		classes = append(classes, &classResolver{classInfo: *class})
	}
	return classes
}
