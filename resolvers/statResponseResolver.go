package resolvers

import "context"

type statResponseResolver struct {
	stat    *statResolver
	message *string
}

func (s *statResponseResolver) Stat(_ context.Context) *statResolver {
	return &statResolver{stat: s.stat.stat}
}

func (s *statResponseResolver) Message(_ context.Context) *string {
	return s.message
}
