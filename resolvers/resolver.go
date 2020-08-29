package resolvers

import (
	"context"
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/services"
	"strconv"
	"strings"
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

func (r *Resolver) ClassChange(ctx context.Context, args struct {
	Id     string
	Class  string
	Weapon *string
}) (*string, error) {
	message, err := r.Services.Adventure.ClassChange(args.Id, strings.Title(strings.ToLower(args.Class)), args.Weapon)
	if err != nil {
		return nil, err
	}
	return message, err
}

func (r *Resolver) ToggleExpEvent(ctx context.Context, args struct{ ExpRate int32 }) (*string, error) {
	err := r.Services.Manage.ToggleExpEvent(int(args.ExpRate))
	message := ""
	if err != nil {
		r.Log.Errorf("error flipping exp flag: %v", err)
		return nil, err
	}
	message = fmt.Sprintf("Successfully flipped exp flag to: %v", args.ExpRate)
	return &message, nil
}

func (r *Resolver) AddLevelTable(ctx context.Context, args struct{ Levels []models.Level }) ([]*levelResolver, error) {
	levelTable, err := r.Services.Manage.CreateExpTable(args.Levels)
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
	id, err := r.Services.Manage.AddNewClass(&args.Class)
	if err != nil {
		r.Log.Errorf("error adding new class: %v", err)
		return nil, err
	}
	return id, nil
}

func (r *Resolver) AddNewUser(ctx context.Context, args struct {
	User   models.User
	Weapon string
}) (*newUserResolver, error) {
	id, message, err := r.Services.Manage.AddNewUser(args.User, args.Weapon)
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

func (r *Resolver) AddNewParty(ctx context.Context, args struct{ Id string }) (*string, error) {
	partyAddMessage, err := r.Services.Adventure.CreateParty(args.Id)
	if err != nil {
		r.Log.Errorf("error adding new party: %v", err)
		return nil, err
	}
	return partyAddMessage, nil
}

func (r *Resolver) JoinParty(ctx context.Context, args struct {
	Id      string
	PartyId string
}) (*string, error) {
	partyMessage, err := r.Services.Adventure.JoinParty(args.PartyId, args.Id)
	if err != nil {
		r.Log.Errorf("error joining the party: %v", err)
		return nil, err
	}
	return partyMessage, nil
}

func (r *Resolver) LeaveParty(ctx context.Context, args struct {
	Id string
}) (*string, error) {
	partyMessage, err := r.Services.Adventure.LeaveParty(args.Id)
	if err != nil {
		r.Log.Errorf("error leaving the party: %v", err)
		return nil, err
	}
	return partyMessage, nil
}

func (r *Resolver) GetAreaList(ctx context.Context) (*[]*areaResolver, error) {
	areaList, err := r.Services.Adventure.GetAreas()
	if err != nil {
		r.Log.Errorf("error getting area list: %v", err)
		return nil, err
	}
	if areaList == nil {
		return nil, nil
	}
	var areas []*areaResolver
	for _, area := range *areaList {
		intValue, err := strconv.ParseInt(area.ID, 10, 32)
		if err != nil {
			r.Log.Errorf("error parsing int from input ID")
			return nil, err
		}
		if intValue < 1000 {
			areas = append(areas, &areaResolver{area: area})
		}
	}
	return &areas, nil
}

func (r *Resolver) GetJobList(ctx context.Context) (*[]*jobClassResolver, error) {
	jobList, err := r.Services.Adventure.GetJobList()
	if err != nil {
		r.Log.Errorf("error getting job list: %v", err)
		return nil, err
	}
	r.Log.Debugf("jobList: %v", jobList)
	var jobClasses []*jobClassResolver
	for _, job := range *jobList {
		if job.LevelRequirement < 150 {
			jobClasses = append(jobClasses, &jobClassResolver{jobClass: job})
		}
	}
	return &jobClasses, nil
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

func (r *Resolver) UpgradeEquipment(ctx context.Context, args struct {
	Id        string
	Equipment string
}) (*string, error) {
	msg, err := r.Services.Adventure.UpdateEquipmentPiece(args.Id, strings.ToLower(args.Equipment))
	if err != nil {
		r.Log.Errorf("error updating equipment piece: %v", err)
		return nil, err
	}
	return msg, nil
}

func (r *Resolver) JobAdvance(ctx context.Context, args struct {
	Id     string
	Class  string
	Weapon string
}) (*string, error) {
	message, err := r.Services.Adventure.ClassAdvance(args.Id, strings.Title(strings.ToLower(args.Weapon)), strings.Title(strings.ToLower(args.Class)), nil)
	if err != nil {
		r.Log.Errorf("error class advancing: %v", err)
		return nil, err
	}
	return message, nil
}

func (r *Resolver) GetUpgradeCost(ctx context.Context, args struct {
	Id        string
	Equipment string
}) (*string, error) {
	message, err := r.Services.Adventure.GetEquipmentPieceCost(args.Id, strings.ToLower(args.Equipment))
	if err != nil {
		r.Log.Errorf("error getting equipment cost: %v", err)
		return nil, err
	}

	return message, nil
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
