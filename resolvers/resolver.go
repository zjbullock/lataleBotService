package resolvers

import (
	"context"
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/services"
	"sort"
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

func (r *Resolver) AddNewBoss(ctx context.Context, args struct{ Boss models.Monster }) (*string, error) {
	id, err := r.Services.Manage.AddNewBoss(args.Boss)
	if err != nil {
		r.Log.Errorf("error adding new boss: %v", err)
		return nil, err
	}
	return id, nil
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

func (r *Resolver) EquipItem(ctx context.Context, args struct{ Id, Name string }) (*string, error) {
	message, err := r.Services.Adventure.EquipItem(args.Id, strings.TrimSpace(args.Name))
	if err != nil {
		r.Log.Errorf("error equipping specified item")
		return nil, err
	}
	return message, err
}

func (r *Resolver) BuyItem(ctx context.Context, args struct{ Id, Name string }) (*string, error) {
	message, err := r.Services.Adventure.BuyItem(args.Id, strings.TrimSpace(args.Name))
	if err != nil {
		r.Log.Errorf("error buying specified item")
		return nil, err
	}
	return message, err
}

func (r *Resolver) SellItem(ctx context.Context, args struct{ Id, Name string }) (*string, error) {
	message, err := r.Services.Adventure.SellItem(args.Id, strings.TrimSpace(args.Name), nil, 1)
	if err != nil {
		r.Log.Errorf("error selling specified item")
		return nil, err
	}
	return message, err
}

func (r *Resolver) SellAllItems(ctx context.Context, args struct{ Id string }) (*string, error) {
	user, message, err := r.Services.Adventure.GetUserInfo(args.Id)
	if err != nil {
		r.Log.Errorf("error getting user info")
		return message, err
	}
	inventory, _, err := r.Services.Adventure.GetUserInventory(args.Id)
	if err != nil {
		r.Log.Errorf("error getting user inventory")
	}
	if inventory != nil {
		for equip, count := range inventory.Equipment {
			_, err := r.Services.Adventure.SellItem(args.Id, strings.TrimSpace(equip), user, count)
			if err != nil {
				r.Log.Errorf("error equipping specified item")
				return nil, err
			}
		}
	}
	finishedSelling := "You've sold all items in your inventory!"
	return &finishedSelling, nil
}

func (r *Resolver) GetUserInventory(ctx context.Context, args struct{ Id string }) (*inventoryResponseResolver, error) {
	inventory, message, err := r.Services.Adventure.GetUserInventory(args.Id)
	if err != nil {
		r.Log.Errorf("error getting user inventory: %v", err)
		return nil, err
	}
	return &inventoryResponseResolver{inventory: inventory, message: message}, nil
}

func (r *Resolver) GetShopInventory(ctx context.Context, args struct{ Id string }) (*[]*itemInfoResolver, error) {
	items, err := r.Services.Adventure.GetShopInventory(args.Id)
	if err != nil {
		r.Log.Errorf("error getting user inventory: %v", err)
		return nil, err
	}
	if items != nil && len(*items) == 0 {
		return nil, nil
	}
	var sortedItems = *items
	sort.Slice(sortedItems, func(i, j int) bool {
		return *sortedItems[i].LevelRequirement > *sortedItems[j].LevelRequirement
	})
	var itemResolvers []*itemInfoResolver
	for _, item := range sortedItems {
		itemResolvers = append(itemResolvers, &itemInfoResolver{item: item})
	}
	return &itemResolvers, nil
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

func (r *Resolver) AddNewParty(ctx context.Context, args struct{ Id string }) (*string, error) {
	partyAddMessage, err := r.Services.Adventure.CreateParty(args.Id)
	if err != nil {
		r.Log.Errorf("error adding new party: %v", err)
		return nil, err
	}
	return partyAddMessage, nil
}

func (r *Resolver) AddNewItem(ctx context.Context, args struct{ Item models.Item }) (*string, error) {
	id, err := r.Services.Manage.AddNewItem(&args.Item)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (r *Resolver) GetItemInfo(ctx context.Context, args struct{ ItemName string }) (*itemResponseResolver, error) {
	itemInfo, message, err := r.Services.Adventure.GetItemInfo(strings.TrimSpace(args.ItemName))
	if err != nil {
		r.Log.Errorf("error retrieving info for the specified item: %v", err)
		return nil, err
	}
	return &itemResponseResolver{item: itemInfo, message: message}, nil
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

func (r *Resolver) KickFromParty(ctx context.Context, args struct {
	Id     string
	KickId string
}) (*string, error) {
	partyMessage, err := r.Services.Adventure.KickParty(args.Id, args.KickId)
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
		jobClasses = append(jobClasses, &jobClassResolver{jobClass: job})
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

func (r *Resolver) ConvertToInventory(_ context.Context) (*string, error) {
	message, err := r.Services.Manage.ConvertToInventorySystemBatch()
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *Resolver) GetBossBonus(ctx context.Context, args struct{ Id string }) (*statResolver, error) {
	bossId, err := strconv.ParseInt(args.Id, 10, 32)
	if err != nil {
		r.Log.Errorf("error parsing input string into int:%v", err)
		return nil, err
	}
	bonus, err := r.Services.Adventure.GetBossBonus(int32(bossId))
	if err != nil {
		r.Log.Errorf("error retrieving boss bonus: %v", err)
		return nil, err
	}
	return &statResolver{stat: &models.StatModifier{
		CriticalRate:           bonus.CriticalRate,
		MaxDPS:                 bonus.MaxDPS,
		MinDPS:                 bonus.MinDPS,
		CriticalDamageModifier: bonus.CriticalDamageModifier,
		Defense:                bonus.Defense,
		Accuracy:               bonus.Accuracy,
		Evasion:                bonus.Evasion,
		HP:                     bonus.HP,
		SkillProcRate:          bonus.SkillProcRate,
		Recovery:               bonus.Recovery,
		SkillDamageModifier:    bonus.SkillDamageModifier,
	}}, nil
}

func (r *Resolver) GetBossList(ctx context.Context, args struct{ Id string }) (*[]string, error) {
	bosses, err := r.Services.Adventure.GetBosses(args.Id)
	if err != nil {
		r.Log.Errorf("error getting bosses: %v", err)
		return nil, err
	}
	return bosses, nil
}

func (r *Resolver) GetUserClassInfo(ctx context.Context, args struct{ Id string }) (*equipmentResolver, error) {
	panic("Implement Me!")
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

func (r *Resolver) GetBossFight(ctx context.Context, args struct {
	BossId string
	UserId string
}) (*adventureResponseResolver, error) {
	adventureLog, message, err := r.Services.Adventure.GetBossBattle(strings.Title(strings.ToLower(args.BossId)), args.UserId)
	if err != nil {
		r.Log.Errorf("error getting adventure log: %v", err)
		return nil, err
	}
	return &adventureResponseResolver{log: adventureLog, message: message}, nil
}
