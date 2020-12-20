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

func (j *jobClassResolver) LevelRequirement(_ context.Context) *float64 {
	var levelRequirement = float64(j.jobClass.LevelRequirement)
	return &levelRequirement
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

func (j *jobClassResolver) Trait(_ context.Context) *traitResolver {
	if j.jobClass.Trait == nil {
		return nil
	}
	return &traitResolver{trait: *j.jobClass.Trait}
}
