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
		Damage    services.Damage
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
	area, _, err := r.Services.Adventure.GetArea(args.Area)
	if err != nil {
		r.Log.Errorf("error getting area info: %v", err)
		return nil, err
	}
	if area == nil {
		status := "Area specified does not exist!"
		return &status, nil
	}
	id, err := r.Services.Manage.AddNewMonster(area, args.Monster)
	if err != nil {
		message := "Unable to add the new monster"
		return &message, err
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
	return &statResponseResolver{stat: stats, message: message}, nil
}

func (r *Resolver) GetUserInfo(ctx context.Context, args struct{ Id string }) (*userResponseResolver, error) {
	user, message, err := r.Services.Adventure.GetUserInfo(args.Id)
	return &userResponseResolver{user: user, message: message}, err
}

func (r *Resolver) GetUserClassInfo(ctx context.Context, args struct{ Id string }) (*equipmentResolver, error) {
	panic("Implement Me!")
}

func (r *Resolver) GetArea(ctx context.Context, args struct{ Id string }) (*areaResponseResolver, error) {
	area, message, err := r.Services.Adventure.GetArea(args.Id)
	if err != nil {
		r.Log.Errorf("error getting area: $v", err)
		return nil, err
	}
	return &areaResponseResolver{areaInfo: area, message: message}, nil
}

func (r *Resolver) GetClassInfo(ctx context.Context, args struct{ Id string }) (*jobClassResponseResolver, error) {
	jobClass, err := r.Services.Adventure.GetJobClassDescription(args.Id)
	var message string
	if err != nil {
		message = "Provided class does not exist!"
	}
	return &jobClassResponseResolver{jobClass: jobClass, message: &message}, nil
}

func (r *Resolver) GetAdventure(ctx context.Context, args struct {
	AreaId string
	UserId string
}) (*adventureResponseResolver, error) {
	adventureLog, message, err := r.Services.Adventure.GetAdventure(args.AreaId, args.UserId)
	if err != nil {
		r.Log.Errorf("error getting adventure log: %v", err)
		return nil, err
	}
	return &adventureResponseResolver{log: adventureLog, message: message}, nil
}
