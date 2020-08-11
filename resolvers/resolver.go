package resolvers

import (
	"context"
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/services"
)

type Resolver struct {
	Services struct {
		Adventure services.Adventure
	}
	Log loggo.Logger
}

func (r *Resolver) AddNewClass(ctx context.Context, args struct{ Class models.JobClass }) (*string, error) {
	id, err := r.Services.Adventure.AddNewClass(args.Class)
	if err != nil {
		r.Log.Errorf("error adding new class: %v", err)
		return nil, err
	}
	return id, nil
}

func (r *Resolver) AddNewUser(ctx context.Context, args struct{ User models.User }) (*string, error) {
	id, err := r.Services.Adventure.AddNewUser(args.User)
	if err != nil {
		r.Log.Errorf("error adding new user: %v", err)
		return nil, err
	}
	return id, nil
}

func (r *Resolver) GetUserBaseStats(ctx context.Context, args struct{ Id string }) (*statResponseResolver, error) {
	stats, message, err := r.Services.Adventure.GetBaseStat(args.Id)
	if err != nil {
		r.Log.Errorf("error getting base stat: %v", err)
		return nil, err
	}
	return &statResponseResolver{stat: &statResolver{stat: *stats}, message: message}, nil
}
