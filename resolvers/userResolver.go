package resolvers

import (
	"context"
	"lataleBotService/models"
)

type userResolver struct {
	user *models.User
}

func (u *userResolver) Name(_ context.Context) string {
	return u.user.Name
}

func (u *userResolver) ID(_ context.Context) string {
	return u.user.ID
}

func (u *userResolver) Ely(_ context.Context) *float64 {
	if u.user.Ely == nil {
		return nil
	}
	ely := float64(*u.user.Ely)
	return &ely
}

func (u *userResolver) CurrentClass(_ context.Context) string {
	return u.user.CurrentClass
}

func (u *userResolver) Classes(_ context.Context) *[]*classResolver {
	var classes []*classResolver
	for _, class := range u.user.ClassMap {
		classes = append(classes, &classResolver{classInfo: *class})
	}
	return &classes
}

func (u *userResolver) PartyMembers(_ context.Context) *[]string {
	return u.user.PartyMembers
}
