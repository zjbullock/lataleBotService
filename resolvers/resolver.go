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

func (r *Resolver) IncreaseLevelCap(ctx context.Context, args struct{ LevelCap int32 }) ([]*levelResolver, error) {
	levelTable, err := r.Services.Manage.IncreaseLevelCap(int(args.LevelCap))
	if err != nil {
		r.Log.Errorf("error getting level table: %v", err)
		return nil, err
	}
	var levels []*levelResolver
	for _, level := range *levelTable {
		levels = append(levels, &levelResolver{level: level})
	}
	return levels, nil
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

func (r *Resolver) AddNewUser(ctx context.Context, args struct{ User models.User }) (*newUserResolver, error) {
	id, message, err := r.Services.Manage.AddNewUser(args.User)
	if err != nil {
		r.Log.Errorf("error adding new user: %v", err)
		return nil, err
	}
	return &newUserResolver{newUserResponse: models.NewUserResponse{ID: id, Message: message}}, nil
}

func (r *Resolver) AddNewEquipmentSheet(ctx context.Context, args struct{ Equipment models.EquipmentSheet }) (*string, error) {
	id, err := r.Services.Manage.AddNewEquipmentSheet(args.Equipment)
	if err != nil {
		r.Log.Errorf("error adding new equipment sheet: %v", err)
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

func (r *Resolver) GetUserInfo(ctx context.Context, args struct{ Id string }) (*userResponseResolver, error) {
	user, message, err := r.Services.Adventure.GetUserInfo(args.Id)
	return &userResponseResolver{user: user, message: message}, err
}

func (r *Resolver) GetUserClassInfo(ctx context.Context, args struct{ Id string }) (*equipmentResolver, error) {
	panic("Implement Me!")
}

func (r *Resolver) GetArea(ctx context.Context, args struct{ Id string }) (*areaResolver, error) {
	area, err := r.Services.Adventure.GetArea(args.Id)
	if err != nil {
		r.Log.Errorf("error getting area: $v", err)
		return nil, err
	}
	return &areaResolver{area: *area}, nil
}
