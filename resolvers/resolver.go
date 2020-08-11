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
		Manage    services.Manage
	}
	Log loggo.Logger
}

func (r *Resolver) AddNewMonster(ctx context.Context, args struct {
	Area    string
	Monster models.Monster
}) (*string, error) {
	area, err := r.Services.Adventure.GetArea(args.Area)
	if err != nil {
		r.Log.Errorf("error getting area info: %v", err)
		return nil, err
	}
	if area == nil {
		status := "area specified does not exist"
		return &status, nil
	}
	id, err := r.Services.Manage.AddNewMonster(args.Area, args.Monster)
	if err != nil {
		return nil, err
	}
	return id, err
}

func (r *Resolver) AddNewArea(ctx context.Context, args struct{ Area models.Area }) (*string, error) {
	id, err := r.Services.Manage.AddNewArea(args.Area)
	if err != nil {
		r.Log.Errorf("error adding new area: %v", err)
		return nil, err
	}
	return id, nil
}

func (r *Resolver) AddNewClass(ctx context.Context, args struct{ Class models.JobClass }) (*string, error) {
	id, err := r.Services.Manage.AddNewClass(args.Class)
	if err != nil {
		r.Log.Errorf("error adding new class: %v", err)
		return nil, err
	}
	return id, nil
}

func (r *Resolver) AddNewUser(ctx context.Context, args struct{ User models.User }) (*string, error) {
	id, err := r.Services.Manage.AddNewUser(args.User)
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

func (r *Resolver) GetArea(ctx context.Context, args struct{ Id string }) (*areaResolver, error) {
	area, err := r.Services.Adventure.GetArea(args.Id)
	if err != nil {
		r.Log.Errorf("error getting area: $v", err)
		return nil, err
	}
	return &areaResolver{area: *area}, nil
}
