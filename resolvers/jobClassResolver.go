package resolvers

import (
	"context"
	"lataleBotService/models"
)

type jobClassResolver struct {
	jobClass models.JobClass
}

func (j *jobClassResolver) Name(_ context.Context) *string {
	return &j.jobClass.Name
}

func (j *jobClassResolver) LevelRequirement(_ context.Context) *int32 {
	return &j.jobClass.LevelRequirement
}

func (j *jobClassResolver) ClassRequirement(_ context.Context) *string {
	return j.jobClass.ClassRequirement
}

func (j *jobClassResolver) Weapons(_ context.Context) *[]models.Weapon {
	return &j.jobClass.Weapons
}

func (j *jobClassResolver) Stats(_ context.Context) *models.StatModifier {
	return &j.jobClass.Stats
}

func (j *jobClassResolver) Description(_ context.Context) *string {
	return &j.jobClass.Description
}
