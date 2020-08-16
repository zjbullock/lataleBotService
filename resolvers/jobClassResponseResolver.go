package resolvers

import (
	"context"
	"lataleBotService/models"
)

type jobClassResponseResolver struct {
	jobClass *models.JobClass
	message  *string
}

func (j *jobClassResponseResolver) JobClassInfo(_ context.Context) *jobClassResolver {
	return &jobClassResolver{jobClass: j.jobClass}
}

func (j *jobClassResponseResolver) Message(_ context.Context) *string {
	return j.message
}
