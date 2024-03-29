package services

import (
	"encoding/json"
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/repositories"
	"lataleBotService/utils"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Adventure interface {
	GetBosses(id string) (*[]string, error)
	KickParty(leaderId, kickId string) (*string, error)
	GetBossBonus(id int32) (*models.BossBonus, error)
	ClassChange(id, class string, weapon *string) (*string, error)
	CreateParty(id string) (*string, error)
	JoinParty(partyId, id string) (*string, error)
	LeaveParty(id string) (*string, error)
	EquipItem(id, item string) (*string, error)
	EquipBestItems(id string) (*string, error)
	BuyItem(id, item string) (*string, error)
	SellItem(id, item string, quantity int, sellBoss bool) (*string, error)
	LockItem(id, item string) (*string, error)
	UnlockItem(id, lockedItem string) (*string, error)
	GetBaseStat(id string) (*models.StatModifier, *string, error)
	ClassAdvance(id, weapon, class string, givenClass *string) (*string, error)
	GetJobList() (*[]models.JobClass, error)
	GetAdventure(areaId, userId string) (*[]string, *string, error)
	GetExpGainRate(id string) (*int, error)
	GetJobClassDescription(id string) (*models.JobClass, error)
	GetArea(id string) (*models.Area, *string, error)
	GetAreas() (*[]models.Area, error)
	GetUserInfo(id string) (*models.User, *string, error)
	GetBossBattle(bossId, userId string) (*[]string, *string, error)
	GetSetBonus(id string) (*models.SetBonus, error)
	GetItemInfo(itemName string) (*models.Item, *string, error)
	GetUserInventory(id string) (*models.Inventory, *string, error)
	GetShopInventory(id string) (*[]models.Item, error)
	GetBankInventory(id string) (*models.Inventory, *string, error)
	BankDeposit(id, item string) (*string, error)
	BankWithdraw(id, item string) (*string, error)
}

type adventure struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	users     repositories.UserRepository
	levels    repositories.LevelRepository
	ascension repositories.AscensionRepository
	equipment repositories.EquipmentRepository
	config    repositories.ConfigRepository
	party     repositories.PartyRepository
	boss      repositories.BossRepository
	item      repositories.ItemRepository
	bank      repositories.BankRepository
	setBonus  repositories.SetBonusRepository
	battle    Battle
	env       map[string]interface{}
	log       loggo.Logger
}

func NewAdventureService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, equips repositories.EquipmentRepository, levels repositories.LevelRepository, ascension repositories.AscensionRepository, config repositories.ConfigRepository, party repositories.PartyRepository, boss repositories.BossRepository, item repositories.ItemRepository, bank repositories.BankRepository, setBonus repositories.SetBonusRepository, env map[string]interface{}, log loggo.Logger) Adventure {
	return &adventure{
		areas:     areas,
		classes:   classes,
		users:     users,
		equipment: equips,
		levels:    levels,
		ascension: ascension,
		config:    config,
		party:     party,
		battle:    NewBattleService(log),
		boss:      boss,
		item:      item,
		bank:      bank,
		setBonus:  setBonus,
		env:       env,
		log:       log,
	}
}

func (a *adventure) GetItemInfo(itemName string) (*models.Item, *string, error) {
	items, err := a.item.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "name",
			Op:    "==",
			Value: itemName,
		},
	})
	if err != nil {
		message := fmt.Sprintf("Unable to find info regarding an item with the name: %s", itemName)
		return nil, &message, nil
	}
	if items == nil || items != nil && len(items) == 0 {
		message := fmt.Sprintf("Unable to find info regarding an item with the name: %s", itemName)
		return nil, &message, nil
	}
	return &items[0], nil, nil
}

func (a *adventure) GetBossBonus(id int32) (*models.BossBonus, error) {
	bosses, err := a.boss.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "bossBonus.id",
			Op:    "==",
			Value: id,
		},
	})
	if err != nil {
		a.log.Errorf("error retrieving boss bonus with id %v: %v", id, err)
		return nil, err
	}
	var bossBonus *models.BossBonus
	for _, boss := range *bosses {
		bossBonus = boss.BossBonus
	}
	return bossBonus, nil
}

func (a *adventure) GetSetBonus(id string) (*models.SetBonus, error) {
	setBonus, err := a.setBonus.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error retrieving set bonus with id %v: %v", id, err)
		return nil, err
	}
	return setBonus, nil
}

func (a *adventure) GetBosses(id string) (*[]string, error) {
	var availableBosses []string
	bosses, err := a.boss.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "level",
			Op:    ">=",
			Value: 0,
		},
	})
	if err != nil {
		a.log.Errorf("error getting bosses: %v", err)
		message := "There was a problem getting the list of available bosses."
		availableBosses = append(availableBosses, message)
		return &availableBosses, nil
	}
	if bosses == nil || len(*bosses) == 0 {
		message := "There are currently no bosses available to fight you."
		availableBosses = append(availableBosses, message)
		return &availableBosses, nil
	}
	for _, boss := range *bosses {
		if boss.AscensionLevel != nil {
			availableBosses = append(availableBosses, fmt.Sprintf("%s	|	%s	|	⭐%v", boss.Name, boss.BossBonus.Name, *boss.AscensionLevel))
		} else if boss.AscensionLevel == nil {
			availableBosses = append(availableBosses, fmt.Sprintf("%s	|	%s	|	%v", boss.Name, boss.BossBonus.Name, boss.Level))
		}
	}
	return &availableBosses, nil
}

func (a *adventure) GetShopInventory(id string) (*[]models.Item, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		return nil, nil
	}
	var items []models.Item
	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		a.log.Errorf("error getting class info: %v", err)
		return nil, nil
	}
	for _, weapon := range classInfo.Weapons {
		weaponItems, err := a.item.QueryDocuments(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "<=",
				Value: user.ClassMap[user.CurrentClass].Level + 10,
			},
			{
				Path:  "shop",
				Op:    "==",
				Value: true,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: weapon.Name,
			},
		})
		if err != nil {
			a.log.Errorf("error getting items for shop: %v", err)
			return nil, nil
		}
		items = append(items, weaponItems...)
	}

	armorItems, err := a.item.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "<=",
			Value: user.ClassMap[user.CurrentClass].Level + 10,
		},
		{
			Path:  "shop",
			Op:    "==",
			Value: true,
		},
		{
			Path:  "type.itemType",
			Op:    "==",
			Value: "armor",
		},
	})
	items = append(items, armorItems...)
	if err != nil {
		a.log.Errorf("error getting items for shop: %v", err)
		return nil, nil
	}
	return &items, nil
}

func (a *adventure) CreateParty(id string) (*string, error) {
	//1.  Ensure that user is currently a player of the bot.
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	//2.  Ensure that user is not currently in a party
	checkIfInParty, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "members",
			Op:    "array-contains",
			Value: id,
		},
	})
	if err != nil {
		message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
		return &message, nil
	}
	if checkIfInParty != nil {
		message := fmt.Sprintf("You are currently in a party!  To leave your current party, first run **/latale party leave**.")
		return &message, nil
	}
	checkIfPartyLeader, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "leader",
			Op:    "==",
			Value: id,
		},
	})
	if err != nil {
		message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
		return &message, nil
	}
	if checkIfPartyLeader != nil {
		message := fmt.Sprintf("You are currently the leader of a party!  To disband your current party, first run **/latale party leave**.")
		return &message, nil
	}
	//3.  If user is not currently in a party, create a new party.
	partyId, err := a.party.InsertDocument(nil, &models.Party{
		Leader:  user.ID,
		Members: []string{user.ID},
	})
	if err != nil {
		a.log.Errorf("error generating party: %v", err)
		message := fmt.Sprintf("A problem was encountered creating a party.  Sorry for the inconvenience")
		return &message, nil
	}
	//4.  Update user doc to create relation between them and their current party
	user.Party = partyId
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error updating user info: %v", err)
		message := fmt.Sprintf("A problem was encountered creating a party.  Sorry for the inconvenience")
		return &message, nil
	}
	message := fmt.Sprintf("**Congratulations**, your party has been created!\nPlease have members join your party by using the \"***/latale party join partycode:%s***\" command to NiceHat.\nTo keep players you do not want to join the party from using the command, please do not post the command publicly", *partyId)
	return &message, nil
}

func (a *adventure) JoinParty(partyId, id string) (*string, error) {
	userInfo, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error retrieving userInfo: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	//1.  Ensure that user is not currently in a party
	checkIfInParty, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "members",
			Op:    "array-contains",
			Value: id,
		},
	})
	if err != nil {
		message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
		return &message, nil
	}
	if checkIfInParty != nil {
		message := fmt.Sprintf("You are currently in a party!")
		return &message, nil
	}
	checkIfPartyLeader, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "leader",
			Op:    "==",
			Value: id,
		},
	})
	if err != nil {
		message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
		return &message, nil
	}
	if checkIfPartyLeader != nil {
		message := fmt.Sprintf("You are currently the leader of this party!")
		return &message, nil
	}

	party, err := a.party.ReadDocument(partyId)
	if err != nil {
		a.log.Errorf("error retrieving party: %v", err)
		message := "The requested party does not exist!"
		return &message, nil
	}
	//2.  If user is not in party, check that requested party limit is not met.
	a.log.Debugf("partyMembers: %v", party.Members)
	if len(party.Members) == 8 {
		message := fmt.Sprintf("The requested party is already full!")
		return &message, nil
	}
	//3.  If not met, add user to the party.
	party.Members = append(party.Members, id)
	_, err = a.party.UpdateDocument(partyId, &party)
	if err != nil {
		a.log.Errorf("error updating party info: %v", err)
		message := fmt.Sprintf("A problem was encountered adding you to the party.  Sorry for the inconvenience")
		return &message, nil
	}
	userInfo.Party = &partyId
	_, err = a.users.UpdateDocument(userInfo.ID, userInfo)
	if err != nil {
		a.log.Errorf("error updating user info: %v", err)
		message := fmt.Sprintf("A problem was encountered creating a party.  Sorry for the inconvenience")
		return &message, nil
	}
	message := fmt.Sprintf("You have successfully been added to the party!  To leave the party in the future, simply run the command **/latale party leave**")
	return &message, nil
}

func (a *adventure) KickParty(leaderId, kickId string) (*string, error) {
	_, err := a.users.ReadDocument(leaderId)
	if err != nil {
		a.log.Errorf("error retrieving userInfo: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	party, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "leader",
			Op:    "==",
			Value: leaderId,
		},
	})
	if err != nil {
		message := fmt.Sprintf("You do not appear to be the leader of a party!")
		return &message, nil
	}
	if party.Leader != leaderId {
		message := fmt.Sprintf("You are not the leader of the party, and do not have kick permissions!")
		return &message, nil
	}
	if party.Leader == kickId {
		message := fmt.Sprintf("You cannot kick yourself from the party!")
		return &message, nil
	}
	partyMemberInfo, err := a.users.ReadDocument(kickId)
	if err != nil {
		a.log.Errorf("error retrieving userInfo: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	var newParty []string
	found := false
	for _, member := range party.Members {
		if member != kickId {
			newParty = append(newParty, member)
		} else {
			found = true
		}
	}
	if !found {
		message := fmt.Sprintf("The party member specified does not seem to be part of your party.")
		return &message, nil
	}
	party.Members = newParty
	_, err = a.party.UpdateDocument(*party.ID, party)
	if err != nil {
		message := fmt.Sprintf("Failed to kick this party member!")
		return &message, nil
	}
	partyMemberInfo.PartyMembers = nil
	partyMemberInfo.Party = nil
	_, err = a.users.UpdateDocument(partyMemberInfo.ID, partyMemberInfo)
	if err != nil {
		message := fmt.Sprintf("Failed to kick this party member!")
		return &message, nil
	}
	message := fmt.Sprintf("<@%s> has been successfully kicked and removed from the party.", kickId)
	return &message, nil
}

func (a *adventure) LeaveParty(id string) (*string, error) {
	userInfo, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error retrieving userInfo: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	//1.  Ensure that user is not currently in a party
	var partyId *string
	checkIfPartyLeader, err := a.party.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "leader",
			Op:    "==",
			Value: id,
		},
	})
	if err != nil {
		message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
		return &message, nil
	}
	if checkIfPartyLeader != nil {
		partyId = checkIfPartyLeader.ID
	}
	if partyId == nil {
		checkIfInParty, err := a.party.QueryDocuments(&[]models.QueryArg{
			{
				Path:  "members",
				Op:    "array-contains",
				Value: id,
			},
		})
		if err != nil {
			message := fmt.Sprintf("A problem was encountered fetching previous party info.  Sorry for the inconvenience")
			return &message, nil
		}
		if checkIfInParty != nil {
			partyId = checkIfInParty.ID
		} else {
			message := fmt.Sprintf("You are not currently in a party!")
			return &message, nil
		}
	}
	party, err := a.party.ReadDocument(*partyId)
	if err != nil {
		a.log.Errorf("error retrieving party: %v", err)
		return nil, err
	}
	message := ""
	if party.Leader == id {
		//Remove relation of all party members to party first
		if party.Members != nil && len(party.Members) > 0 {
			for _, member := range party.Members {
				userInfo, err := a.users.ReadDocument(member)
				if err != nil {
					a.log.Errorf("error retrieving userInfo: %v", err)
					message := "User has not created an account yet."
					return &message, nil
				}
				//Remove the party's id from the user's info and update document
				userInfo.Party = nil
				_, err = a.users.UpdateDocument(member, userInfo)
				if err != nil {
					a.log.Errorf("error updating userInfo: %v", err)
					message := "There was a problem removing you from the party!"
					return &message, nil
				}
			}
		}
		//Delete party from party collection
		err := a.party.DeleteDocument(*party.ID)
		if err != nil {
			a.log.Errorf("error disbanding the party: %v", err)
			return nil, err
		}
		message = "You have disbanded the party!"
	} else {
		//Create array to hold party members left after leaving
		partyLeftover := []string{}
		for _, member := range party.Members {
			if member != id {
				partyLeftover = append(partyLeftover, member)
			}
		}
		party.Members = partyLeftover
		//Update the party to remove the player
		_, err = a.party.UpdateDocument(*party.ID, &party)
		if err != nil {
			a.log.Errorf("error saving new party document: %v", err)
			message := "There was a problem removing you from the party!"
			return &message, nil
		}
		//Remove the party's id from the user's info and update document
		userInfo.Party = nil
		_, err = a.users.UpdateDocument(id, userInfo)
		if err != nil {
			a.log.Errorf("error updating userInfo: %v", err)
			message := "There was a problem removing you from the party!"
			return &message, nil
		}
		message = "You have left the party!"
	}
	return &message, nil
}

func (a *adventure) GetArea(id string) (*models.Area, *string, error) {
	area, err := a.areas.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting area: %v", err)
		message := "Unable to get area with that name!"
		return nil, &message, err
	}
	return area, nil, nil
}

func (a *adventure) ClassChange(id, class string, weapon *string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		message := "User has not created an account yet."
		return &message, nil
	}
	if user.CurrentClass == class {
		message := fmt.Sprintf("You're already a %s!", user.CurrentClass)
		return &message, nil
	}
	if user.ClassMap[class] != nil && weapon == nil {
		user.CurrentClass = user.ClassMap[class].Name
		_, err := a.users.UpdateDocument(user.ID, user)
		if err != nil {
			a.log.Errorf("There was an error processing the class change request: %v", err)
			message := fmt.Sprintf("There was an error processing the class change request.")
			return &message, nil
		}
		message := fmt.Sprintf("%s has successfully class changed to %s", user.Name, class)
		return &message, nil
	} else if user.ClassMap[class] != nil && weapon != nil {
		message := fmt.Sprintf("You do not need to specify a weapon when changing class to a class you have previously changed to.")
		return &message, nil
	}
	classInfo, err := a.classes.ReadDocument(class)
	if err != nil {
		a.log.Errorf("error getting current class: %v", err)
		message := fmt.Sprintf("The %s class does not exist.  Please select a valid class with a valid weapon", class)
		return &message, nil
	}
	if classInfo.ClassRequirement != nil {
		for _, weaponInfo := range classInfo.Weapons {
			if _, ok := user.ClassMap[*classInfo.ClassRequirement]; ok {
				if weaponInfo.Name == *user.ClassMap[*classInfo.ClassRequirement].Equipment.Weapon.Type.WeaponType {
					newWeapon := weaponInfo.Name
					weapon = &newWeapon
				}
			}
		}
	}

	if user.ClassMap[class] == nil && weapon != nil {
		currentWeapon := strings.Title(strings.ToLower(*weapon))
		if classInfo.Tier == 1 {
			for _, classWeapon := range classInfo.Weapons {
				if classWeapon.Name == strings.Title(strings.ToLower(*weapon)) {
					startingWeapon, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: currentWeapon,
						},
					})
					if err != nil {
						panic("error getting weapons")
					}
					top, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: "Top",
						},
					})
					if err != nil {
						panic("error getting tops")
					}
					bottom, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: "Bottom",
						},
					})
					if err != nil {
						panic("error getting tops")
					}
					headpiece, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: "Headpiece",
						},
					})
					if err != nil {
						panic("error getting headpieces")
					}
					gloves, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: "Gloves",
						},
					})
					if err != nil {
						panic("error getting gloves")
					}
					boots, err := a.item.QueryForDocument(&[]models.QueryArg{
						{
							Path:  "levelRequirement",
							Op:    "==",
							Value: 1,
						},
						{
							Path:  "type.weaponType",
							Op:    "==",
							Value: "Boots",
						},
					})
					if err != nil {
						panic("error getting boots")
					}
					user.ClassMap[class] = &models.ClassInfo{
						Name:  classInfo.Name,
						Level: classInfo.LevelRequirement,
						Exp:   0,
						Equipment: models.Equipment{
							Weapon:    *startingWeapon,
							Top:       *top,
							Headpiece: *headpiece,
							Bottom:    *bottom,
							Glove:     *gloves,
							Shoes:     *boots,
						},
						SetBonuses: make(map[string]*models.SetBonus),
					}
					user.CurrentClass = class
					_, err = a.users.UpdateDocument(user.ID, user)
					if err != nil {
						a.log.Errorf("There was an error processing the class change request: %v", err)
						message := fmt.Sprintf("There was an error processing the class change request.")
						return &message, nil
					}
					message := fmt.Sprintf("%s has successfully class changed to %s", user.Name, class)
					return &message, nil
				}
			}
			message := fmt.Sprintf("%s is not a valid weapon for this class!", strings.Title(strings.ToLower(*weapon)))
			return &message, nil
		} else {
			fmt.Printf("userClasses: %v\n", user.ClassMap)
			a.log.Infof("weapon, %v", currentWeapon)
			for _, learnedClass := range user.ClassMap {
				if learnedClass.Name == *classInfo.ClassRequirement && learnedClass.Level >= classInfo.LevelRequirement {
					a.log.Infof("weapon, %s", currentWeapon)
					message, err := a.ClassAdvance(id, strings.Title(strings.ToLower(currentWeapon)), class, &learnedClass.Name)
					if err != nil {
						a.log.Errorf("error while switching and upgrading a class: %v", err)
						return nil, err
					}
					return message, nil
				}
			}
			message := fmt.Sprintf("You do not meet one or more requirements to change to this job.  Ensure you have the proper level on the required class.")
			return &message, nil
		}
	}
	message := "Please provide a weapon, as this is your first time creating this class, and you do not currently have the required weapon type equipped."
	return &message, nil
}

func (a *adventure) GetExpGainRate(id string) (*int, error) {
	expGainRate, err := a.config.ReadDocument("exp")
	if err != nil {
		return nil, err
	}
	return expGainRate["exp"], nil
}

func (a *adventure) GetAreas() (*[]models.Area, error) {
	areaList, err := a.areas.QueryDocuments(nil)
	if err != nil {
		a.log.Errorf("error querying for area list: %v", err)
		return nil, err
	}
	return areaList, nil
}

func (a *adventure) GetBaseStat(id string) (*models.StatModifier, *string, error) {
	//1.  Get User Data based on ID
	a.log.Debugf("id: %s", id)
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		message := "User has not created an account yet."
		return nil, &message, nil
	}
	a.log.Debugf("user: %v", user)
	class, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		a.log.Errorf("error reading currently selected class")
		return nil, nil, err
	}
	currentStats, err := a.calculateBaseStat(*user, class.Stats)
	if err != nil {
		a.log.Errorf("error getting base stats: %v", err)
		message := fmt.Sprintf("There was an issue getting your base stats.")
		return nil, &message, err
	}

	return currentStats, nil, nil
}

func (a *adventure) GetJobClassDescription(id string) (*models.JobClass, error) {
	jobClass, err := a.classes.ReadDocument(strings.Title(strings.ToLower(id)))
	if err != nil {
		a.log.Errorf("Job :%s doesn't exist.", id)
		return nil, err
	}
	return jobClass, nil
}

func (a *adventure) ClassAdvance(id, weapon, class string, givenClass *string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		message := "You have not created an account yet!"
		return &message, nil
	}
	classInfo, err := a.classes.ReadDocument(class)
	if err != nil {
		message := fmt.Sprintf("The class: %v, does not exist!", class)
		return &message, nil
	}
	if user.ClassMap[classInfo.Name] != nil {
		message := fmt.Sprintf("You've already advanced to %s.", classInfo.Name)
		return &message, nil
	}
	if classInfo.Tier < 2 {
		message := fmt.Sprintf("The specified class is a First Tier Class, and cannot be advanced to!")
		return &message, nil
	}
	a.log.Debugf("ClassRequirement: %s", *classInfo.ClassRequirement)
	for _, wep := range classInfo.Weapons {
		if wep.Name == weapon {
			if *classInfo.ClassRequirement == user.CurrentClass || givenClass != nil && *classInfo.ClassRequirement == *givenClass {
				classToUse := user.CurrentClass
				if givenClass != nil {
					classToUse = *givenClass
				}
				if classInfo.LevelRequirement <= user.ClassMap[classToUse].Level {
					equips := user.ClassMap[classToUse].Equipment
					if classInfo.Tier != 3 {
						equips = *a.determineStartingGear(classInfo.Tier, &user.ClassMap[classToUse].Equipment, weapon)
					}
					user.ClassMap[class] = &models.ClassInfo{
						Name:        classInfo.Name,
						Level:       user.ClassMap[classToUse].Level,
						Exp:         user.ClassMap[classToUse].Exp,
						Equipment:   equips,
						BossBonuses: user.ClassMap[classToUse].BossBonuses,
					}
					user.CurrentClass = classInfo.Name
					_, err := a.users.UpdateDocument(user.ID, user)
					if err != nil {
						a.log.Errorf("error updating user doc with new class: %v", err)
						return nil, err
					}
					message := fmt.Sprintf("**Congratulations, %v, on your advancement to %v!**\n", user.Name, user.CurrentClass)
					message += a.jobTierMessages(classInfo.Tier)
					return &message, nil
				}
				message := fmt.Sprintf("You do not meet the level requirement of %v to complete this job advancement!", classInfo.LevelRequirement)
				return &message, nil
			}
			message := fmt.Sprintf("You do not meet the class requirement of %s to complete this job advancement!", *classInfo.ClassRequirement)
			return &message, nil
		}
	}
	message := fmt.Sprintf("The specified weapon does not exist on this job!")
	return &message, nil
}

func (a *adventure) jobTierMessages(tier int32) string {
	if tier == 2 {
		message := fmt.Sprintf("Congratulations on reaching a Second Tier Class, you have obtained the ability to equip the following items: **Bindi, Glasses, Earrings, Ring, Cloak, and Stockings**.\n")
		message += fmt.Sprintf("Your continued patronage is appreciated.\n")
		return message
	} else if tier == 3 {
		message := fmt.Sprintf("Congratulations on reaching a Third Tier Class.  Third Tier Classes posses powerful passives known as Traits.  They can give your character an edge in progressing through content.\n")
		message += fmt.Sprintf("Your continued patronage is appreciated.\n")
		return message
	} else if tier == 4 {
		message := fmt.Sprintf("Congratulations on reaching a Fourth Tier Class.  Fourth Tier Classes expand further upon their previous traits and later gain access to class specific equipment.\n")
		message += fmt.Sprintf("Your continued patronage is appreciated.\n")
		return message
	}
	return ""
}

func (a *adventure) determineStartingGear(tier int32, currentEquips *models.Equipment, weaponType string) *models.Equipment {
	equipLevel := 50
	if tier == 4 {
		equipLevel = 140
	}
	var weapon *models.Item
	a.log.Infof("weaponType, %s", weaponType)
	if weaponType == *currentEquips.Weapon.Type.WeaponType && tier > 2 {
		a.log.Infof("Type.WeaponType: %s", *currentEquips.Weapon.Type.WeaponType)
		weapon = &currentEquips.Weapon
	} else {
		newWeapon, err := a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: weaponType,
			},
		})
		if err != nil {
			panic("error getting weapon!")
		}
		weapon = newWeapon
	}
	top, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: equipLevel,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: "Top",
		},
	})
	if err != nil {
		panic("error getting top!")
	}
	bottom, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: equipLevel,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: "Bottom",
		},
	})
	if err != nil {
		panic("error getting bottom!")
	}
	headpiece, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: equipLevel,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: "Headpiece",
		},
	})
	if err != nil {
		panic("error getting headpiece!")
	}
	gloves, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: equipLevel,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: "Gloves",
		},
	})
	if err != nil {
		panic("error getting gloves!")
	}
	boots, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: equipLevel,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: "Boots",
		},
	})
	if err != nil {
		panic("error getting boots!")
	}
	var bindi = currentEquips.Bindi
	var glasses = currentEquips.Glasses
	var earring = currentEquips.Earring
	var ring = currentEquips.Ring
	var mantle = currentEquips.Cloak
	var stocking = currentEquips.Stockings

	if bindi == nil {
		bindi, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Bindi",
			},
		})
		if err != nil {
			panic("error getting bindi!")
		}
		glasses, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Glasses",
			},
		})
		if err != nil {
			panic("error getting glasses!")
		}
		earring, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Earrings",
			},
		})
		if err != nil {
			panic("error getting earrings!")
		}
		ring, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Ring",
			},
		})
		if err != nil {
			panic("error getting ring!")
		}
		mantle, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Cloak",
			},
		})
		if err != nil {
			panic("error getting mantle!")
		}
		stocking, err = a.item.QueryForDocument(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: equipLevel,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: "Stockings",
			},
		})
		if err != nil {
			panic("error getting stocking!")
		}
	}

	return &models.Equipment{
		Weapon:    *weapon,
		Top:       *top,
		Headpiece: *headpiece,
		Bottom:    *bottom,
		Glove:     *gloves,
		Shoes:     *boots,
		Bindi:     bindi,
		Glasses:   glasses,
		Earring:   earring,
		Ring:      ring,
		Cloak:     mantle,
		Stockings: stocking,
	}
}

//func (a *adventure) UpdateEquipmentPiece(id, equipment string) (*string, error) {
//	if equipment == "gloves" {
//		equipment = "glove"
//	}
//	if equipment == "earring" {
//		equipment = "earrings"
//	}
//	if equipment == "cloak" {
//		equipment = "mantle"
//	}
//	//1.  Get User Info
//	user, err := a.users.ReadDocument(id)
//	if err != nil {
//		return nil, err
//	}
//
//	//2.  Based on equipment piece, pass current gear level to ProcessUpgrade
//	message, err := a.processUpgrade(user, strings.ToLower(equipment))
//	if err != nil {
//		a.log.Errorf("error processing upgrade: %v", err)
//		return nil, err
//	}
//	return message, nil
//}

//func (a *adventure) GetEquipmentPieceCost(id, equipment string) (*string, error) {
//	if equipment == "gloves" {
//		equipment = "glove"
//	}
//	if equipment == "earring" {
//		equipment = "earrings"
//	}
//	if equipment == "cloak" {
//		equipment = "mantle"
//	}
//	user, err := a.users.ReadDocument(id)
//	if err != nil {
//		return nil, err
//	}
//	var equipmentInterface map[string]interface{}
//	equips := user.ClassMap[user.CurrentClass].OldEquipmentSheet
//	bytes, _ := json.Marshal(&equips)
//	json.Unmarshal(bytes, &equipmentInterface)
//	equip := equipmentInterface[equipment]
//	a.log.Debugf("equips :%v", equip)
//	if equip == nil {
//		message := fmt.Sprintf("%s is not a valid piece of equipment or you have not unlocked this slot yet!", equipment)
//		return &message, nil
//	}
//	oldEquipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)))
//	if err != nil {
//		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
//		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
//		return &message, nil
//	}
//	equipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)+1.0))
//	if err != nil {
//		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
//		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
//		return &message, nil
//	}
//	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
//	if err != nil {
//		message := "Something happened while trying to get class info..."
//		return &message, nil
//	}
//	if classInfo.Tier < equipSheet.TierRequirement {
//		message := "You are not at the proper class advancement to advance this piece of equipment any further!"
//		return &message, nil
//	}
//
//	message := ""
//	switch equipment {
//	case "bindi":
//		oldBindi := strconv.FormatFloat(oldEquipSheet.BindiHP, 'f', -1, 64)
//		bindi := strconv.FormatFloat(equipSheet.BindiHP, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("HP gained from bindi: **%s** -> **%s**\n", oldBindi, bindi)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "glasses":
//		oldGlasses := strconv.FormatFloat(oldEquipSheet.GlassesCritDamage*100.0, 'f', -1, 64)
//		glasses := strconv.FormatFloat(equipSheet.GlassesCritDamage*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("Critical battle gained from glasses: **%s%%** -> **%s%%**\n", oldGlasses, glasses)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "earrings", "earring":
//		oldEarrings := strconv.FormatFloat(oldEquipSheet.EarringCritRate*100.0, 'f', -1, 64)
//		earrings := strconv.FormatFloat(equipSheet.EarringCritRate*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("Critical Rate gained from earrings: **%s%%** -> **%s%%**\n", oldEarrings, earrings)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "ring":
//		oldRing := strconv.FormatFloat(oldEquipSheet.RingCritRate*100.0, 'f', -1, 64)
//		ring := strconv.FormatFloat(equipSheet.RingCritRate*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("Critical Rate gained from ring: **%s%%** -> **%s%%**\n", oldRing, ring)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "cloak", "mantle":
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("battle gained from cloak: **%1.f** -> **%1.f**\n", oldEquipSheet.MantleDamage, equipSheet.MantleDamage)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "stockings":
//		oldStockingsEvasion := strconv.FormatFloat(oldEquipSheet.StockingEvasion*100.0, 'f', -1, 64)
//		stockingEvasion := strconv.FormatFloat(equipSheet.StockingEvasion*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
//		message += fmt.Sprintf("Evasion gained from stockings: **%s%%** -> **%s%%**\n", oldStockingsEvasion, stockingEvasion)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "weapon":
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
//		message += fmt.Sprintf("battle gained from weapon: **%1.f** -> **%1.f**\n", oldEquipSheet.WeaponDPS, equipSheet.WeaponDPS)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "shoe", "shoes", "boot", "boots":
//		oldShoeEvasion := strconv.FormatFloat(oldEquipSheet.ShoeEvasion*100.0, 'f', -1, 64)
//		shoeEvasion := strconv.FormatFloat(equipSheet.ShoeEvasion*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
//		message += fmt.Sprintf("Evasion gained from shoes: **%s%%** -> **%s%%**\n", oldShoeEvasion, shoeEvasion)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "glove", "gloves":
//		oldGloveAccuracy := strconv.FormatFloat(oldEquipSheet.GloveAccuracy*100.0, 'f', -1, 64)
//		gloveAccuracy := strconv.FormatFloat(equipSheet.GloveAccuracy*100.0, 'f', -1, 64)
//		oldGloveCritDamage := strconv.FormatFloat(oldEquipSheet.GloveCriticalDamage*100.0, 'f', -1, 64)
//		gloveCritDamage := strconv.FormatFloat(equipSheet.GloveCriticalDamage*100.0, 'f', -1, 64)
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
//		message += fmt.Sprintf("Accuracy gained from gloves: **%s%%** -> **%s%%**\n", oldGloveAccuracy, gloveAccuracy)
//		message += fmt.Sprintf("Critical battle gained from gloves: **%s%%** -> **%s%%**\n", oldGloveCritDamage, gloveCritDamage)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	case "body":
//		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
//		message += fmt.Sprintf("Defense gained from body armor: **%1.f** -> **%1.f**\n", oldEquipSheet.ArmorDefense, equipSheet.ArmorDefense)
//		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
//	}
//	return &message, nil
//}

//
//func (a *adventure) processUpgrade(user *models.User, equipment string) (*string, error) {
//	var equipmentInterface map[string]interface{}
//	equips := user.ClassMap[user.CurrentClass].OldEquipmentSheet
//	a.log.Debugf("equipment: %v", equipment)
//	bytes, _ := json.Marshal(&equips)
//	json.Unmarshal(bytes, &equipmentInterface)
//	equip := equipmentInterface[equipment]
//	a.log.Debugf("equips :%v", equip)
//	if equip == nil {
//		message := fmt.Sprintf("%s is not a valid piece of equipment, or you have not unlocked this slot yet!", equipment)
//		return &message, nil
//	}
//
//	//3.  Check if an upgrade is available.  If no doc found, gear does not exist, or something happened.
//	equipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)+1.0))
//	if err != nil {
//		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
//		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
//		return &message, nil
//	}
//
//	if equipSheet.LevelRequirement > user.ClassMap[user.CurrentClass].Level {
//		message := fmt.Sprintf("You do not meet the level requirement to upgrade this piece of gear!  Required Level: %v", equipSheet.LevelRequirement)
//		return &message, nil
//	}
//	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
//	if err != nil {
//		message := "Something happened while trying to get class info..."
//		return &message, nil
//	}
//	if classInfo.Tier < equipSheet.TierRequirement {
//		message := "You are not at the proper class advancement to advance this piece of equipment any further!"
//		return &message, nil
//	}
//
//	currentEquipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)))
//	if err != nil {
//		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
//		return nil, err
//	}
//	//If the cost is met, decrease user ely by cost, and upgrade specified piece of equipment.
//
//	s := ""
//	cost := int64(0)
//	switch equipment {
//	case "glove", "gloves", "body", "shoe", "shoes", "weapon":
//		cost = equipSheet.Cost
//		break
//	default:
//		cost = equipSheet.AccessoryCost
//	}
//	if *user.Ely >= cost {
//		currentValue := equipmentInterface[equipment].(float64)
//		currentValue++
//		equipmentInterface[equipment] = currentValue
//		bytes, _ = json.Marshal(&equipmentInterface)
//		var newEquips models.OldEquipmentSystem
//		json.Unmarshal(bytes, &newEquips)
//		a.log.Debugf("newEquips :%v", newEquips)
//		userClass := user.ClassMap[user.CurrentClass]
//		userClass.OldEquipmentSheet = &newEquips
//		ely := *user.Ely
//		ely -= cost
//		user.Ely = &ely
//		user.ClassMap[user.CurrentClass] = userClass
//		_, err := a.users.UpdateDocument(user.ID, user)
//		if err != nil {
//			a.log.Errorf("error updating user document: %v", err)
//			return nil, err
//		}
//		if equipment == "weapon" {
//			s = fmt.Sprintf("Successfully upgraded %s from %s to %s!", equipment, currentEquipSheet.WeaponMap[user.ClassMap[user.CurrentClass].CurrentWeapon], equipSheet.WeaponMap[user.ClassMap[user.CurrentClass].CurrentWeapon])
//		} else {
//			s = fmt.Sprintf("Successfully upgraded %s from %s to %s!", equipment, currentEquipSheet.Name, equipSheet.Name)
//		}
//	} else {
//		s = fmt.Sprintf("Insufficient Ely!  You need %v more Ely to complete this upgrade.", cost-*user.Ely)
//		return &s, nil
//	}
//	return &s, nil
//}

func (a *adventure) GetUserInfo(id string) (*models.User, *string, error) {
	a.log.Debugf("id: %s", id)
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		s := "user with this id not found"
		return nil, &s, nil
	}
	if user.Party != nil {
		party, err := a.party.ReadDocument(*user.Party)
		if err != nil {
			a.log.Errorf("error fetching party info")
			return user, nil, nil
		}
		var partyMemberInfo []string
		partyMemberInfo = append(partyMemberInfo, fmt.Sprintf("\n**Party Leader:**"))
		memberInfo, err := a.users.ReadDocument(party.Leader)
		if err != nil {
			a.log.Errorf("error getting party member info: %v", memberInfo)
			return user, nil, nil
		}
		partyMemberInfo = append(partyMemberInfo, fmt.Sprintf("**Name:** %s, **Class:** %s, **Level:** %v", memberInfo.Name, memberInfo.CurrentClass, memberInfo.ClassMap[memberInfo.CurrentClass].Level))

		partyMemberInfo = append(partyMemberInfo, fmt.Sprintf("\n**Party Members:**"))
		for _, member := range party.Members {
			memberInfo, err := a.users.ReadDocument(member)
			if err != nil {
				a.log.Errorf("error getting party member info: %v", memberInfo)
				return user, nil, nil
			}
			partyMemberInfo = append(partyMemberInfo, fmt.Sprintf("**Name:** %s, **Class:** %s, **Level:** %v", memberInfo.Name, memberInfo.CurrentClass, memberInfo.ClassMap[memberInfo.CurrentClass].Level))
		}
		user.PartyMembers = &partyMemberInfo
	}
	return user, nil, nil
}

func (a *adventure) getEquipmentMap(classEquips *models.OldEquipmentSystem) (map[string]*models.OldEquipmentSheet, error) {
	classEquipmentMap := make(map[string]*models.OldEquipmentSheet)
	//Determine Top
	equipmentSheetBody, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Body))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with provided equipment")
		return nil, err
	}
	classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBody)
	//Determine Gloves
	if classEquipmentMap[strconv.Itoa(classEquips.Glove)] == nil {
		equipmentSheetGloves, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Glove))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetGloves)
	}
	//Determine Shoes
	if classEquipmentMap[strconv.Itoa(classEquips.Shoes)] == nil {
		equipmentSheetShoes, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Shoes))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetShoes)
	}
	//Determine WeaponMap
	if classEquipmentMap[strconv.Itoa(classEquips.Weapon)] == nil {
		equipmentSheetWeapon, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Weapon))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetWeapon)
	}

	//Determine Bindi
	if classEquips.Bindi != nil && classEquipmentMap[strconv.Itoa(*classEquips.Bindi)] == nil {

		equipmentSheetBindi, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Bindi))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBindi)
	}

	//Determine Glasses
	if classEquips.Glasses != nil && classEquipmentMap[strconv.Itoa(*classEquips.Glasses)] == nil {

		equipmentSheetGlasses, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Glasses))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetGlasses)
	}

	//Determine Earring
	if classEquips.Earring != nil && classEquipmentMap[strconv.Itoa(*classEquips.Earring)] == nil {

		equipmentSheetEarring, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Earring))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetEarring)
	}

	//Determine Ring
	if classEquips.Ring != nil && classEquipmentMap[strconv.Itoa(*classEquips.Ring)] == nil {

		equipmentSheetRing, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Ring))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetRing)
	}

	//Determine Cloak
	if classEquips.Cloak != nil && classEquipmentMap[strconv.Itoa(*classEquips.Cloak)] == nil {

		equipmentSheetMantle, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Cloak))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetMantle)
	}

	//Determine Stockings
	if classEquips.Stockings != nil && classEquipmentMap[strconv.Itoa(*classEquips.Stockings)] == nil {

		equipmentSheetStockings, err := a.equipment.ReadDocument(strconv.Itoa(*classEquips.Stockings))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetStockings)
	}

	return classEquipmentMap, nil
}

func (a *adventure) GetJobList() (*[]models.JobClass, error) {
	jobs, err := a.classes.QueryDocuments(nil)
	if err != nil {
		a.log.Errorf("error getting list of jobs: %v", err)
		return nil, err
	}
	return jobs, err
}

func (a *adventure) GetAdventure(areaId, userId string) (*[]string, *string, error) {
	/*
		1.  Pull User Current stats
		2.  Pull Area Monster list where -1 <= monsterLevel - userLevel <= 3
		4.  Separate monsters into map, with rank as key
		5.  Randomly Generate value from 1-100 (Value represents chances of encountering a certain rank, with ranks being 1-3 and encounters being a 60%,35%,5% chance respectively.  If the rank does not appear in the monster list, it rounds downward).
		6.  Begin combat, with player having priority.  Roll first to hit (userAcc - enemyEva)
		7.  If hits, roll to determine if user successfully used a skill, crit, or both.
		8.  Perform damageCalculations
		9.  Repeat same steps this time for monster(s)
		10.  Recover user and monster health based on recovery %.
		11.  Loop until combat is finished.
		12.  If user successfully defeats the enemies, then updateUser class doc with exp, ely, and level changes.
		13.  If user dies, do nothing.
		14.  Restore user health to max at the end of the combat.
		15.  Return log of events.
	*/
	user, err := a.users.ReadDocument(userId)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return nil, &message, nil
	}
	area, err := a.areas.ReadDocument(areaId)
	if err != nil {
		a.log.Errorf("error getting area info: %v", err)
		message := "Could not find an area with that code.  Please be sure to use the codes specified in `/latale adventure arealist`."
		return nil, &message, nil
	}
	if user.Party != nil {
		adventureLog, err := a.createPartyAdventureLog(user, area)
		if err != nil {
			a.log.Errorf("encountered error generating adventure log: %v", err)
			return &adventureLog, nil, err
		}
		return &adventureLog, nil, nil
	}
	var monsterMap = make(map[string]*[]models.Monster)
	for _, monster := range area.Monsters {
		if monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] == nil {
			monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] = &[]models.Monster{}
		}
		updatedList := *monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)]
		updatedList = append(updatedList, monster)
		monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] = &updatedList
	}

	a.log.Debugf("monsters possible: %v", monsterMap)
	randSource := rand.NewSource(time.Now().UnixNano())
	rarityGenerator := rand.New(randSource)
	monsters := a.determineMonsterRarity(monsterMap, rarityGenerator)
	if monsters == nil {
		afraid := fmt.Sprintf("The monsters in the %s are too afraid of fighting %s", areaId, userId)
		return &[]string{afraid}, nil, nil
	}
	monster := a.determineMonster(*monsters, rarityGenerator)
	currentStats, _, err := a.GetBaseStat(userId)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		message := fmt.Sprintf("Unable to get %s's base stats!", user.Name)
		return nil, &message, err
	}
	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		a.log.Errorf("error getting user class info: %v", err)
		message := fmt.Sprintf("Unable to get class info for %s", user.Name)
		return nil, &message, err
	}

	adventureLog, err := a.createAdventureLog([]*models.UserBlob{{User: user, JobClass: classInfo, BaseStats: currentStats, BattleStats: currentStats, MaxHP: int(currentStats.HP), HitCount: 1, Buffs: make(map[string]models.Buff), Debuffs: make(map[string]models.CrowdControlTrait)}}, models.MonsterBlob{Monster: monster, Name: monster.Name, Bound: false, Exp: monster.Exp, Ely: monster.Ely, Buffs: make(map[string]models.Buff), CurrentHP: int32(monster.Stats.HP), StatModifier: &monster.Stats, BattleStats: &monster.Stats, Debuffs: make(map[string]models.CrowdControlTrait), Rank: monster.Rank}, area.DropRange)
	if err != nil {
		a.log.Errorf("encountered error generating adventure log: %v", err)
		return &adventureLog, nil, err
	}
	return &adventureLog, nil, nil
}

func (a *adventure) createPartyAdventureLog(user *models.User, area *models.Area) ([]string, error) {
	var adventureLog []string
	logLine, onCooldown := a.checkAdventureCooldown(user, false)
	if onCooldown {
		adventureLog = append(adventureLog, *logLine)
		return adventureLog, nil
	}
	partyMemberInfos, err := a.generatePartyBlob(user)
	if err != nil {
		a.log.Errorf("error generating party blob: %v", err)
		return nil, err
	}
	randSource := rand.NewSource(time.Now().UnixNano())
	encounteredMonsters, err := a.generateMonsterBlob(area, randSource, len(partyMemberInfos))
	if err != nil {
		a.log.Errorf("error generating monster blob: %v", err)
		return nil, err
	}
	adventureLog, err = a.partyBattleLog(partyMemberInfos, encounteredMonsters, adventureLog, user, area.DropRange)
	if err != nil {
		a.log.Errorf("error creating adventurelog: %v", err)
		return nil, err
	}
	return adventureLog, nil
}

func (a *adventure) generateMonsterBlob(area *models.Area, randSource rand.Source, partyMemberInfos int) ([]*models.MonsterBlob, error) {
	monsterCount := 1
	if partyMemberInfos > 1 {
		randomMonsterCount := rand.New(randSource)
		monsterCount = randomMonsterCount.Intn(partyMemberInfos) + 1
	}
	var encounteredMonsters []*models.MonsterBlob
	rarityGenerator := rand.New(randSource)
	var monsterMap = make(map[string]*[]models.Monster)
	for _, monster := range area.Monsters {
		if monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] == nil {
			monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] = &[]models.Monster{}
		}
		updatedList := *monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)]
		updatedList = append(updatedList, monster)
		monsterMap[utils.ThirtyTwoBitIntToString(monster.Rank)] = &updatedList
	}
	a.log.Debugf("monsters possible: %v", monsterMap)
	for i := 0; i < monsterCount; i++ {
		monsters := a.determineMonsterRarity(monsterMap, rarityGenerator)
		monster := a.determineMonster(*monsters, rarityGenerator)
		monsterRank := ""
		for i := int32(0); i < monster.Rank; i++ {
			monsterRank += "!"
		}
		encounteredMonsters = append(encounteredMonsters, &models.MonsterBlob{
			CurrentHP:    int32(monster.Stats.HP),
			StatModifier: &monster.Stats,
			BattleStats:  &monster.Stats,
			Name:         monster.Name + " " + string('A'+i) + monsterRank,
			Ely:          monster.Ely,
			Exp:          monster.Exp,
			Buffs:        make(map[string]models.Buff),
			Debuffs:      make(map[string]models.CrowdControlTrait),
		})
	}
	return encounteredMonsters, nil
}

func (a *adventure) GetUserInventory(id string) (*models.Inventory, *string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return nil, &message, nil
	}
	return &user.Inventory, nil, nil
}

func (a *adventure) GetBankInventory(id string) (*models.Inventory, *string, error) {
	bank, err := a.bank.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user bank: %v", err)
		if bank == nil {
			_, err := a.bank.InsertDocument(id, &models.Inventory{
				Equipment: make(map[string]int),
				Consume:   make(map[string]int),
				Event:     make(map[string]int),
			})
			if err != nil {
				a.log.Errorf("error getting bank for user: %s", id)
				return nil, nil, err
			}
			return &models.Inventory{
				Equipment: make(map[string]int),
				Consume:   make(map[string]int),
				Event:     make(map[string]int),
			}, nil, nil
		}
	}

	return bank, nil, nil
}

func (a *adventure) BuyItem(id, item string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return &message, nil
	}
	itemData, err := a.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "name",
			Op:    "==",
			Value: item,
		},
		{
			Path:  "shop",
			Op:    "==",
			Value: true,
		},
	})
	if err != nil {
		a.log.Errorf("error getting item info: %v", err)
		message := "There was a problem finding an item with that name."
		return &message, nil
	}
	if itemData == nil {
		message := fmt.Sprintf("Sorry, that item was unable to be found in the shop...")
		return &message, nil
	}
	if *user.Ely-int64(*itemData.Cost) < 0 {
		message := fmt.Sprintf("You do not have enough funds to complete the purchase of the %s", itemData.Name)
		return &message, nil
	}
	if user.Inventory.Equipment == nil {
		user.Inventory.Equipment = make(map[string]int)
		user.Inventory.Event = make(map[string]int)
		user.Inventory.Consume = make(map[string]int)
	}
	user.Inventory.Equipment[itemData.Name]++
	ely := *user.Ely
	ely -= int64(*itemData.Cost)
	user.Ely = &ely
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error buying the item: %v", err)
		message := fmt.Sprintf("Something happened while purchasing the %s.  Purchase failed.", item)
		return &message, nil
	}
	message := fmt.Sprintf("Successfully purchased the %s and added it to your inventory!", item)
	return &message, nil
}

func (a *adventure) SellItem(id, item string, quantity int, sellBoss bool) (*string, error) {
	if strings.Contains(item, "[LOCKED]") {
		message := "Locked items cannot be sold!"
		return &message, nil
	}
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return &message, nil
	}
	itemData, message, err := a.GetItemInfo(item)
	if err != nil {
		a.log.Errorf("error getting item info: %v", err)
		return message, err
	}
	if itemData == nil {
		message := fmt.Sprintf("Sorry, that item was unable to be found in your bag.  Please use the item's name with proper captilization and spacing.")
		return &message, nil
	}
	if (itemData.Boss != nil || int32(*itemData.LevelRequirement) >= user.ClassMap[user.CurrentClass].Level) && sellBoss == false {
		message := fmt.Sprintf("Skipping this item in sellAll as it is a boss item.")
		return &message, nil
	}
	if user.Inventory.Equipment == nil {
		user.Inventory.Equipment = make(map[string]int)
		user.Inventory.Event = make(map[string]int)
		user.Inventory.Consume = make(map[string]int)
	}
	if user.Inventory.Equipment != nil && user.Inventory.Equipment[itemData.Name] > 0 {
		if quantity > user.Inventory.Equipment[itemData.Name] {
			quantity = user.Inventory.Equipment[itemData.Name]
		}
		elyMade := int64(0)
		for i := 0; i < quantity; i++ {
			ely := *user.Ely
			if itemData.Cost != nil {
				ely += int64(*itemData.Cost / 2)
				elyMade += int64(*itemData.Cost / 2)
			}
			user.Inventory.Equipment[itemData.Name]--
			if user.Inventory.Equipment[itemData.Name] == 0 {
				delete(user.Inventory.Equipment, itemData.Name)
			}
			user.Ely = &ely
		}
		_, err := a.users.UpdateDocument(user.ID, user)
		if err != nil {
			a.log.Errorf("error selling updating user: %v", err)
			message := fmt.Sprintf("There was a problem selling your item...")
			return &message, nil
		}
		message := fmt.Sprintf("Successfully sold **%v** ***%s***(s) for ***%v*** ely!", quantity, itemData.Name, elyMade)
		return &message, nil
	}
	noItemMessage := fmt.Sprintf("There does not appear to be an item with that name in your inventory!")
	return &noItemMessage, nil
}

func (a *adventure) EquipBestItems(id string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return &message, nil
	}
	if len(user.Inventory.Equipment) < 1 {
		message := fmt.Sprintf("You are already wearing your best equipment.")
		return &message, nil
	}
	equipmentItems := user.Inventory.Equipment
	var itemList []models.Item
	for item := range equipmentItems {
		if !strings.Contains(item, "[LOCKED]") {
			queriedItem, err := a.item.QueryForDocument(&[]models.QueryArg{
				{
					"name",
					"==",
					item,
				},
			})
			if err != nil {
				message := fmt.Sprintf("Encounted error querying for item data: %v", err)
				return &message, err
			}
			if queriedItem == nil {
				message := fmt.Sprintf("There was a problem retrieving data on an item in your inventory....")
				return &message, err
			}
			itemList = append(itemList, *queriedItem)
		}
	}
	userClass := user.ClassMap[user.CurrentClass]
	var equipMap map[string]interface{}
	currentUserEquipment, _ := json.Marshal(userClass.Equipment)
	json.Unmarshal(currentUserEquipment, &equipMap)
	for _, item := range itemList {
		if item.RequiredClasses != nil && len(*item.RequiredClasses) > 0 {
			found := false
			for _, class := range *item.RequiredClasses {
				if user.CurrentClass == *class {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if float64(userClass.Level) >= *item.LevelRequirement {
			var jsonBody []byte
			if item.Type.Type == "weapon" {
				jsonBody, err = json.Marshal(equipMap["weapon"])
			} else {
				jsonBody, err = json.Marshal(equipMap[strings.ToLower(*item.Type.WeaponType)])
			}
			if err != nil {
				a.log.Errorf("attempting to marshal item info resulted in: %v", err)
				return nil, err
			}
			var currentItem models.Item
			json.Unmarshal(jsonBody, &currentItem)

			if item.Type.Type == "armor" && (*item.LevelRequirement > *currentItem.LevelRequirement || item.AscensionLevelRequirement != nil && currentItem.AscensionLevelRequirement == nil || item.AscensionLevelRequirement != nil && currentItem.AscensionLevelRequirement != nil && *item.AscensionLevelRequirement > *currentItem.AscensionLevelRequirement) {

				newUser, message := a.changeEquippedItem(&item, user)
				if message != nil {
					a.log.Infof(*message)
				}
				user = newUser
			} else if item.Type.Type == "weapon" && (*item.LevelRequirement > *currentItem.LevelRequirement || item.AscensionLevelRequirement != nil && currentItem.AscensionLevelRequirement == nil || item.AscensionLevelRequirement != nil && currentItem.AscensionLevelRequirement != nil && *item.AscensionLevelRequirement > *currentItem.AscensionLevelRequirement) {
				newUser, message := a.changeEquippedItem(&item, user)
				if message != nil {
					a.log.Infof(*message)
				}
				user = newUser
			}
		}
	}
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error updating user with new equip: %v", err)
		failMessage := fmt.Sprintf("There was an error equipping the new piece of equipment.")
		return &failMessage, nil
	}
	message := fmt.Sprintf("Successfully equipped user with the highest level equipment found in their inventory!")
	return &message, nil
}

func (a *adventure) BankDeposit(id, item string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("problem reading userInfo: %v", err)
		message := "You must first choose a class and get started!"
		return &message, nil
	}
	if user.Inventory.Equipment != nil && user.Inventory.Equipment[item] == 0 {
		message := "You do not appear to possess a %s in your inventory!"
		return &message, nil
	}
	bank, err := a.bank.ReadDocument(id)
	if err != nil {
		newEquipInventory := make(map[string]int)
		newEquipInventory[item]++
		_, err := a.bank.InsertDocument(id, &models.Inventory{
			Equipment: newEquipInventory,
			Consume:   make(map[string]int),
			Event:     make(map[string]int),
		})
		if err != nil {
			message := "no bank found, attempted to create one, but failed"
			return &message, nil
		}
		message := fmt.Sprintf("Successfully inserted **%v** **%s**(s) into your bank inventory!", 1, item)
		return &message, nil
	}
	if len(bank.Equipment) < 45 {
		bank.Equipment[item]++
		_, err = a.bank.UpdateDocument(id, bank)
		if err != nil {
			a.log.Errorf("error updating bank for %s: %v", id, err)
			return nil, err
		}
		user.Inventory.Equipment[item]--
		if user.Inventory.Equipment[item] == 0 {
			delete(user.Inventory.Equipment, item)
		}
		_, err := a.users.UpdateDocument(id, user)
		if err != nil {
			a.log.Errorf("error updating user %s with: %v", id, err)
			return nil, err
		}
		message := fmt.Sprintf("Successfully inserted **%v** **%s**(s) into your bank inventory!", 1, item)
		return &message, nil
	}
	message := "Maximum bank capacity of 45 unique items reached!"
	return &message, nil
}

func (a *adventure) BankWithdraw(id, item string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("problem reading userInfo: %v", err)
		message := "You must first choose a class and get started!"
		return &message, nil
	}
	bank, err := a.bank.ReadDocument(id)
	if err != nil {
		_, err := a.bank.InsertDocument(id, &models.Inventory{
			Equipment: make(map[string]int),
			Consume:   make(map[string]int),
			Event:     make(map[string]int),
		})
		if err != nil {
			message := "no bank found, attempted to create one, but failed"
			return &message, nil
		}
		message := "Created a bank, as you have not opened one yet!"
		return &message, nil
	}
	if bank.Equipment[item] == 0 {
		message := fmt.Sprintf("You do not possess a %s in your bank.", item)
		return &message, nil
	}
	if len(user.Inventory.Equipment) < 45 {
		bank.Equipment[item]--
		if bank.Equipment[item] == 0 {
			delete(bank.Equipment, item)
		}
		_, err = a.bank.UpdateDocument(id, bank)
		if err != nil {
			a.log.Errorf("error updating bank for %s: %v", id, err)
			return nil, err
		}
		user.Inventory.Equipment[item]++
		_, err := a.users.UpdateDocument(id, user)
		if err != nil {
			a.log.Errorf("error updating user %s with: %v", id, err)
			return nil, err
		}
		message := fmt.Sprintf("Successfully inserted **%v** **%s**(s) into your inventory!", 1, item)
		return &message, nil
	}
	message := "Maximum inventory capacity of 45 unique items reached!"
	return &message, nil
}

func (a *adventure) LockItem(id, item string) (*string, error) {
	if strings.Contains(item, "[LOCKED]") {
		message := "A locked item cannot be locked again!"
		return &message, nil
	}
	user, err := a.users.ReadDocument(id)

	if user.Inventory.Equipment[item] < 1 {
		message := fmt.Sprintf("The item %s does not appear to be in your inventory...", item)
		return &message, nil
	}
	lockedItem := item + "[LOCKED]"
	user.Inventory.Equipment[lockedItem]++
	user.Inventory.Equipment[item]--
	if user.Inventory.Equipment[item] == 0 {
		delete(user.Inventory.Equipment, item)
	}
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error updating user with new locked item: %v", err)
		failMessage := fmt.Sprintf("There was an error locking the item.")
		return &failMessage, nil
	}
	successMessage := fmt.Sprintf("Successfully locked 1 %s!", item)
	return &successMessage, nil
}

func (a *adventure) UnlockItem(id, lockedItem string) (*string, error) {
	unLockedItem := lockedItem
	if !strings.Contains(lockedItem, "[LOCKED]") {
		lockedItem = lockedItem + "[LOCKED]"
	} else {
		unLockedItem = strings.ReplaceAll(unLockedItem, "[LOCKED]", "")
	}
	user, err := a.users.ReadDocument(id)
	if user.Inventory.Equipment[lockedItem] < 1 {
		message := fmt.Sprintf("The lockedItem %s does not appear to be in your inventory...", lockedItem)
		return &message, nil
	}
	user.Inventory.Equipment[unLockedItem]++
	user.Inventory.Equipment[lockedItem]--
	if user.Inventory.Equipment[lockedItem] == 0 {
		delete(user.Inventory.Equipment, lockedItem)
	}
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error updating user with new unlocked lockedItem: %v", err)
		failMessage := fmt.Sprintf("There was an error unlocking the lockedItem.")
		return &failMessage, nil
	}
	successMessage := fmt.Sprintf("Successfully unlocked 1 %s!", lockedItem)
	return &successMessage, nil
}

func (a *adventure) EquipItem(id, item string) (*string, error) {
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return &message, nil
	}
	if strings.Contains(item, "[LOCKED]") {
		message := fmt.Sprintf("The %s can't be equipped because it is currently locked.  Unlock it first.", item)
		return &message, nil
	}
	if user.Inventory.Equipment[item] < 1 {
		message := fmt.Sprintf("A %s does not appear to be in your inventory...", item)
		return &message, nil
	}
	items, err := a.item.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "name",
			Op:    "==",
			Value: item,
		},
	})
	if err != nil {
		a.log.Errorf("error getting item with that name: %v", err)
		return nil, err
	}
	if items == nil || items != nil && len(items) == 0 {
		message := "Unable to find an item with that name.  Please check that the spelling on the name."
		return &message, nil
	}
	equipment := items[0]
	if equipment.Type.Type != "weapon" && equipment.Type.Type != "armor" {
		message := fmt.Sprintf("That is a not a valid piece of equipment!  Ensure spelling and capitalization is correct.")
		return &message, err
	}
	if user.ClassMap[user.CurrentClass].Level < int32(*equipment.LevelRequirement) {
		message := fmt.Sprintf("User Level too low to equip the item.  Required Level: %v", int64(*equipment.LevelRequirement))
		return &message, nil
	}
	if equipment.AscensionLevelRequirement != nil && user.AscensionLevel < int32(*equipment.AscensionLevelRequirement) {
		message := fmt.Sprintf("User Ascension Level too low to equip the item.  Required Level: :star:%v", int64(*equipment.AscensionLevelRequirement))
		return &message, nil
	}
	if items[0].RequiredClasses != nil && len(*items[0].RequiredClasses) > 0 {
		found := false
		for _, class := range *items[0].RequiredClasses {
			if user.CurrentClass == *class {
				found = true
			}
		}
		if !found {
			message := "You are not currently one of the required classes for this item."
			return &message, nil
		}
	}
	user, message := a.changeEquippedItem(&equipment, user)
	if message != nil {
		return message, nil
	}
	//Add previously equipped item to inventory
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("error updating user with new equip: %v", err)
		failMessage := fmt.Sprintf("There was an error equipping the new piece of equipment.")
		return &failMessage, nil
	}
	successMessage := fmt.Sprintf("Successfully equipped %s!  Your previously equipped item is now in your inventory.", item)
	return &successMessage, nil
}

func (a *adventure) changeEquippedItem(newItem *models.Item, user *models.User) (*models.User, *string) {
	userClass := user.ClassMap[user.CurrentClass]
	class, err := a.classes.ReadDocument(userClass.Name)
	var equipMap map[string]interface{}
	data, _ := json.Marshal(userClass.Equipment)
	json.Unmarshal(data, &equipMap)
	if err != nil {
		a.log.Errorf("error retrieving user's class info: %v", err)
		message := fmt.Sprintf("An issue was encountered equipping the item.")
		return nil, &message
	}
	a.log.Debugf("equipMap: %v", equipMap)
	if newItem.Type.Type == "weapon" {

		validWeapon := false
		for _, weapon := range class.Weapons {
			if *newItem.Type.WeaponType == weapon.Name {
				validWeapon = true
			}
		}
		if validWeapon {
			user = a.swapItemFromInventoryForAnother(user, newItem, equipMap)
			return user, nil
		} else {
			message := fmt.Sprintf("Your class cannot equip weapons of type: %s", *newItem.Type.WeaponType)
			return user, &message
		}
	}
	user = a.swapItemFromInventoryForAnother(user, newItem, equipMap)
	return user, nil
}

func (a *adventure) swapItemFromInventoryForAnother(user *models.User, newItem *models.Item, equipMap map[string]interface{}) *models.User {
	user.Inventory.Equipment[newItem.Name]--
	if user.Inventory.Equipment[newItem.Name] == 0 {
		delete(user.Inventory.Equipment, newItem.Name)
	}
	item := models.Item{}
	if newItem.Type.Type == "weapon" {
		oldItemData, _ := json.Marshal(equipMap["weapon"])
		json.Unmarshal(oldItemData, &item)
		equipMap["weapon"] = newItem
	} else {
		oldItemData, _ := json.Marshal(equipMap[strings.ToLower(*newItem.Type.WeaponType)])
		json.Unmarshal(oldItemData, &item)
		equipMap[strings.ToLower(*newItem.Type.WeaponType)] = *newItem
	}
	var newEquipment models.Equipment
	newEquipmentData, _ := json.Marshal(equipMap)
	json.Unmarshal(newEquipmentData, &newEquipment)
	user.ClassMap[user.CurrentClass].Equipment = newEquipment
	if newItem.SetBonusId != nil {
		setBonuses := user.ClassMap[user.CurrentClass].SetBonuses
		setBonus, err := a.setBonus.ReadDocument(*newItem.SetBonusId)
		if err != nil {
			a.log.Errorf("error retrieving setbonus: %v", err)
			return user
		}
		if setBonus != nil {
			if setBonuses == nil {
				setBonuses = make(map[string]*models.SetBonus)
				setBonuses[setBonus.Id] = setBonus
			} else if setBonuses[setBonus.Id] != nil {
				oldBonus := setBonuses[setBonus.Id]
				safeBonus := *setBonus.Bonus
				oldBonus.Bonus = &safeBonus
				setBonuses[setBonus.Id] = oldBonus
			} else {
				setBonuses[setBonus.Id] = setBonus
			}
			setBonuses[setBonus.Id].CurrentlyEquipped++
			user.ClassMap[user.CurrentClass].SetBonuses = setBonuses
		}
	}
	user.Inventory.Equipment[item.Name]++
	if item.SetBonusId != nil {
		setBonuses := user.ClassMap[user.CurrentClass].SetBonuses
		setBonuses[*item.SetBonusId].CurrentlyEquipped--
		if setBonuses[*item.SetBonusId].CurrentlyEquipped == int32(0) {
			delete(setBonuses, *item.SetBonusId)
		}
		user.ClassMap[user.CurrentClass].SetBonuses = setBonuses
	}
	return user
}

func (a *adventure) generatePartyBlob(user *models.User) ([]*models.UserBlob, error) {
	party, err := a.party.ReadDocument(*user.Party)
	if err != nil {
		a.log.Errorf("encountered error generating adventure log: %v", err)
		return nil, err
	}
	var partyMemberInfos []*models.UserBlob
	for _, member := range party.Members {
		userInfo, err := a.users.ReadDocument(member)
		if err != nil {
			a.log.Errorf("error getting user info: %v", err)
			return nil, err
		}
		userStats, _, err := a.GetBaseStat(member)
		if err != nil {
			a.log.Errorf("error getting user base stats: %v", err)
			return nil, err
		}
		userJob, err := a.classes.ReadDocument(userInfo.CurrentClass)
		if err != nil {
			a.log.Errorf("error getting current class: %v", err)
			return nil, err
		}
		currentWeapon := userInfo.ClassMap[userInfo.CurrentClass].Equipment.Weapon.Type.WeaponType
		battleStats := *userStats
		partyMemberInfos = append(partyMemberInfos, &models.UserBlob{
			User:        userInfo,
			BaseStats:   userStats,
			BattleStats: &battleStats,
			JobClass:    userJob,
			CurrentHP:   int(userStats.HP),
			MaxHP:       int(userStats.HP),
			UserLevel:   userInfo.ClassMap[userInfo.CurrentClass].Level,
			Weapon:      *currentWeapon,
			Buffs:       make(map[string]models.Buff),
			Debuffs:     make(map[string]models.CrowdControlTrait),
			HitCount:    1,
		})
	}
	return partyMemberInfos, nil
}

func (a *adventure) checkAdventureCooldown(user *models.User, boss bool) (*string, bool) {
	var lastAction time.Time
	var commandType string
	if a.env["environment"] != "PR" {
		return nil, false
	}
	if boss {
		lastAction = user.LastBossActionTime.Add(20 * time.Minute)
		commandType = "boss"
	} else {
		lastAction = user.LastActionTime.Add(2 * time.Minute)
		commandType = "adventure"
	}
	if !time.Now().After(lastAction) {
		timeDifference := lastAction.Sub(time.Now())
		a.log.Debugf("timeDifference: %v", timeDifference)
		minutes := 0
		seconds := 0
		a.log.Debugf("timeDifferenceSeconds: %v", int(timeDifference.Seconds()))
		if int(timeDifference.Seconds()) < 60 {
			seconds = int(timeDifference.Seconds())
		} else {
			for i := 60; int(timeDifference.Seconds())-i >= 0; i += 60 {
				minutes++
				seconds = int(timeDifference.Seconds()) - i
			}
		}
		coolDown := fmt.Sprintf("__**%s**__ must wait **%v** ***Minutes*** and **%v** ***Seconds*** before using the %s command again!", user.Name, minutes, seconds, commandType)
		return &coolDown, true
	}
	return nil, false
}

func (a *adventure) GetBossBattle(bossId, userId string) (*[]string, *string, error) {
	user, err := a.users.ReadDocument(userId)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not yet selected a class, or created an account"
		return nil, &message, nil
	}
	if user.Party == nil {
		message := "You must be in a party to participate in Boss Fights!  Join a party, or create one using `/latale party create`."
		return nil, &message, nil
	}
	boss, err := a.boss.ReadDocument(bossId)
	if err != nil {
		a.log.Errorf("error getting area info: %v", err)
		message := "Could not find a boss with that name.  Please be sure to use the names specified in `/latale boss list`."
		return nil, &message, nil
	}
	partyMembers, err := a.generatePartyBlob(user)
	if err != nil {
		a.log.Errorf("error generating party blob: %v", err)
	}
	var adventureLog []string
	coolDownLog := ""
	for _, partyMember := range partyMembers {
		logLine, coolDown := a.checkAdventureCooldown(partyMember.User, true)
		if coolDown {
			coolDownLog += *logLine + "\n"
		}
		if boss.Level > partyMember.UserLevel {
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s level is lower than the required level of: %v", partyMember.User.Name, boss.Level))
		}
		if boss.AscensionLevel != nil && *boss.AscensionLevel > partyMember.User.AscensionLevel {
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s level is lower than the required ascension level of:  :star:%v", partyMember.User.Name, *boss.AscensionLevel))
		}
	}
	if len(coolDownLog) > 0 {
		adventureLog = append(adventureLog, coolDownLog)
	}
	if len(adventureLog) != 0 {
		return &adventureLog, nil, nil
	}
	adventureLog, err = a.bossBattleLog(partyMembers, models.MonsterBlob{Monster: *boss, Name: boss.Name, Ely: boss.Ely, Exp: boss.Exp, StatModifier: &boss.Stats, BattleStats: &boss.Stats, Buffs: make(map[string]models.Buff), Debuffs: make(map[string]models.CrowdControlTrait)}, user.ID)
	if err != nil {
		a.log.Errorf("error while generating boss log: %v", err)
		return nil, nil, err
	}
	return &adventureLog, nil, nil
}

func (a *adventure) determineBossBonusDrop(userName string, userClassInfo models.ClassInfo, boss models.Monster, adventureLog []string) (models.ClassInfo, []string) {
	dropChance := rand.Float64()
	if userClassInfo.BossBonuses[boss.Name] == nil && dropChance <= *boss.BossBonus.BossDropChance {
		if userClassInfo.BossBonuses == nil {
			userClassInfo.BossBonuses = make(map[string]*models.BossBonus)
		}
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained the boss bonus __**%s**__ for defeating __**%s**__ on their current class!", userName, boss.BossBonus.Name, boss.Name))
		userClassInfo.BossBonuses[boss.Name] = boss.BossBonus
	}
	return userClassInfo, adventureLog
}

func (a *adventure) determineMonsterRarity(monsterMap map[string]*[]models.Monster, rarityGenerator *rand.Rand) *[]models.Monster {
	rarityPercent := rarityGenerator.Intn(100) + 1
	if monsterMap["3"] != nil && rarityPercent >= 96 {
		return monsterMap["3"]
	} else if monsterMap["2"] != nil && rarityPercent <= 95 && rarityPercent >= 60 {
		return monsterMap["2"]
	}
	return monsterMap["1"]
}

func (a *adventure) determineMonster(monsters []models.Monster, rarityGenerator rand.Source) models.Monster {
	monsterSelection := rand.New(rarityGenerator)
	if len(monsters) == 1 {
		return monsters[0]
	}
	monster := monsterSelection.Intn(int(len(monsters)))
	return monsters[monster]
}

func (a *adventure) createAdventureLog(users []*models.UserBlob, monster models.MonsterBlob, dropRange models.LevelRange) ([]string, error) {
	var adventureLog []string
	logLine, onCooldown := a.checkAdventureCooldown(users[0].User, false)
	if onCooldown {
		adventureLog = append(adventureLog, *logLine)
		return adventureLog, nil
	}
	battleWin := false
	userMaxHP := int(users[0].BaseStats.HP)
	monsterMaxHp := int(monster.Monster.Stats.HP)
	currentHP := int(users[0].BaseStats.HP)
	monsterHP := int(monster.Monster.Stats.HP)
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	rankExclamation := ""
	for i := int32(0); i < monster.Rank; i++ {
		rankExclamation += "!"
	}
	adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ has encountered a __**%s**__ **%s**", users[0].User.Name, monster.Name, rankExclamation))
	users, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.BATTLESTARTTRAIT, true)
	adventureLog = append(adventureLog, buffString)
	battleStats := *users[0].BattleStats
	userLevel := users[0].User.ClassMap[users[0].User.CurrentClass].Level
	userWeapon := users[0].User.ClassMap[users[0].User.CurrentClass].Equipment.Weapon.Type.WeaponType
	summons := users[0].Summons
	jobClass := users[0].JobClass
	for currentHP != 0 && monsterHP != 0 {
		if jobClass.Trait != nil && jobClass.Trait.Type == globals.SUMMONTRAIT && jobClass.Trait.ActivationRate != nil {
			summonChance := rand.Float64()

			if summonChance <= *jobClass.Trait.ActivationRate {
				summonLogs, newSummons := a.battle.GenerateSummons(users[0].User.Name, summons, *jobClass.Trait.Summon, battleStats)
				adventureLog = append(adventureLog, summonLogs)
				summons = newSummons
			}
		}
		userLog, damage, statusAilment := a.battle.DetermineHit(randGenerator, users[0].User.Name, monster.Name, battleStats, monster.Monster.Stats, userWeapon, users[0].JobClass, &userLevel, false, users[0].HitCount)
		if len(summons) > 0 {
			for i, summon := range summons {
				summonLog, summonDamage, _ := a.battle.DetermineHit(randGenerator, summon.Name+" "+strconv.Itoa(i+1), monster.Name, summon.StatModifier, monster.Monster.Stats, userWeapon, users[0].JobClass, &userLevel, false, users[0].HitCount)
				userLog += "\n" + summonLog
				damage += summonDamage
			}
		}
		monsterHP = ((int(monsterHP) - int(damage)) + int(math.Abs(float64(monsterHP-damage)))) / 2
		if jobClass.Trait != nil && jobClass.Trait.Type == globals.AFTERATTACKTRAIT && jobClass.Trait.ActivationRate != nil {
			activationChance := rand.Float64()
			if jobClass.Trait.CrowdControl != nil && activationChance <= *jobClass.Trait.ActivationRate {
				userLog += fmt.Sprintf("__**%s**__ activated their trait, __**%s**__!\n", users[0].User.Name, jobClass.Trait.Name)
				if jobClass.Trait.CrowdControl.Type == "drain" {
					monsterHP = int(float64(monsterHP) * 0.975)
					drainHp := int(float64(userMaxHP) * 0.2)
					if currentHP+drainHp > userMaxHP {
						currentHP = userMaxHP
					} else {
						currentHP += drainHp
					}
					userLog += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", users[0].User.Name, drainHp)
					userLog += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", users[0].User.Name, currentHP, userMaxHP)
					userMaxHP += drainHp
				}
			}
		}
		if damage > 0 {
			userLog += fmt.Sprintf("\n__**%s**__'s HP: %v/%v\n", monster.Name, monsterHP, monsterMaxHp)
		}
		if statusAilment != nil {
			monsterDebuffLogs, newMonster := a.battle.InflictStatusAilmentMonster(&monster, *statusAilment)
			userLog += monsterDebuffLogs
			monster = *newMonster
			if statusAilment.Bind {
				monster.Bound = true
			}
		}
		buffedUsers, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.REACTIVETRAIT, false)
		users[0] = buffedUsers[0]
		battleStats = *buffedUsers[0].BattleStats
		userLog += buffString + "\n"
		if monsterHP <= 0 {
			userLog += fmt.Sprintf("__**%s**__ **has successfully defeated the** __**%s**__!\n", users[0].User.Name, monster.Name)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", userLog))
			battleWin = true
			break
		}
		for ailment, debuff := range monster.Debuffs {
			if ailment == "poison" || ailment == "bleed" || ailment == "burn" {
				damageOvertime := 0.0
				switch ailment {
				case "poison":
					damageOvertime = float64(monster.Monster.Stats.HP) * 0.05
					break
				case "burn":
					damageOvertime = float64(monster.Monster.Stats.HP) * 0.025
					break
				case "bleed":
					damageOvertime = float64(monsterHP) * 0.1
					break
				}
				monsterHP = (monsterHP - int(damageOvertime)) + int(math.Abs(float64(monsterHP)-damageOvertime))/2
				userLog += fmt.Sprintf("**%s lost %v HP!** due to **%s**.\n", monster.Name, int(damageOvertime), ailment)
				userLog += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", monster.Name, monsterHP, int(monster.Monster.Stats.HP))
				if monsterHP <= 0 {
					userLog += fmt.Sprintf("__**%s**__ **has successfully defeated the** __**%s**__!\n", users[0].User.Name, monster.Name)
					adventureLog = append(adventureLog, fmt.Sprintf("%s", userLog))
					battleWin = true
					break
				}
			}
			if debuff.Bind {
				monster.Bound = true
			}
			monsterDebuffLogs, newMonster := a.battle.DecreaseMonsterDebuffDuration(&monster)
			if newMonster != nil {
				monster = *newMonster
			}
			if monsterDebuffLogs != "" {
				userLog += monsterDebuffLogs + "\n"
			}
		}
		adventureLog = append(adventureLog, fmt.Sprintf("%s", userLog))
		if !monster.Bound {
			monsterLog, damage, _ := a.battle.DetermineHit(randGenerator, monster.Name, users[0].User.Name, monster.Monster.Stats, battleStats, nil, nil, nil, false, 1)
			currentHP = ((int(currentHP) - int(damage)) + int(math.Abs(float64(currentHP-damage)))) / 2
			monsterLog = "				" + monsterLog
			if damage > 0 {
				monsterLog += fmt.Sprintf("\n				__**%s**__'s HP: %v/%v\n", users[0].User.Name, currentHP, userMaxHP)
				if len(summons) > 0 {
					var aliveSummons []models.Summons
					for i, summon := range summons {
						if summons[i].Duration == nil {
							summons[i].StatModifier.HP -= float64(damage)
							monsterLog += fmt.Sprintf("\n				 __**%s**__ has also taken **%v** damage!\n", summon.Name+" "+strconv.Itoa(i+1), int(damage))
							if summons[i].StatModifier.HP > 0.0 {
								monsterLog += fmt.Sprintf("				__**%s**__'s remaining HP: **%v**\n", summon.Name+" "+strconv.Itoa(i+1), int(summons[i].StatModifier.HP))
								aliveSummons = append(aliveSummons, summons[i])
							} else {
								monsterLog += fmt.Sprintf("__**%s**__'s has died!\n", summon.Name+" "+strconv.Itoa(i+1))
							}
						} else {
							aliveSummons = append(aliveSummons, summons[i])
						}
					}
					summons = aliveSummons
				}
			} else {
				monsterLog += "\n"
			}
			if currentHP <= 0 {
				if jobClass.Trait != nil && jobClass.Trait.Type == globals.DEATHTRAIT && *jobClass.Trait.UsageCount > 0 {
					buffedUsers, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.DEATHTRAIT, false)
					monsterLog += buffString
					userMaxHP = int(buffedUsers[0].BattleStats.HP)
					currentHP = userMaxHP
					skillUsage := *jobClass.Trait.UsageCount
					skillUsage--
					jobClass.Trait.UsageCount = &skillUsage
				} else {
					monsterLog += fmt.Sprintf("				**%s was killed by %s!**\n", users[0].User.Name, monster.Name)
					adventureLog = append(adventureLog, fmt.Sprintf("%s", monsterLog))
					break
				}
			}
			adventureLog = append(adventureLog, fmt.Sprintf("%s", monsterLog))
		} else {
			adventureLog = append(adventureLog, fmt.Sprintf("				**%s** can't move due to being bound!\n", monster.Name))
			monster.Bound = false
		}
		if len(users[0].Buffs) > 0 {
			buffLogs, changedUser := a.battle.DecreaseUserBuffDuration(users[0])
			if changedUser != nil {
				users[0] = changedUser
				battleStats = *users[0].BattleStats
			}
			if buffLogs != "" {
				userLog += buffLogs
			}
		}
		if users[0].JobClass.Trait != nil {
			users[0].MaxHP = userMaxHP
			users[0].CurrentHP = currentHP
			users, buffString = a.battle.DetermineTraitActivations(users, adventureLog, globals.REACTIVETRAIT, false)
			users, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.HEALTRAIT, false)
			adventureLog = append(adventureLog, buffString)
			battleStats = *users[0].BattleStats
			userMaxHP = int(battleStats.HP)
		}
		if len(summons) > 0 {
			decreaseSummonLog, newSummons := a.battle.DecreaseSummonDuration(users[0].User.Name, summons)
			adventureLog = append(adventureLog, decreaseSummonLog)
			summons = newSummons
		}
		adventureLog = append(adventureLog)
		userHeal := int(float64(users[0].BattleStats.HP) * battleStats.Recovery)
		if battleStats.Recovery > 0.0 && currentHP != int(battleStats.HP) {
			healLogs := ""
			if userHeal+currentHP > int(battleStats.HP) {
				currentHP = int(battleStats.HP)
			} else {
				currentHP += userHeal
			}
			healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", users[0].User.Name, userHeal)
			healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", users[0].User.Name, currentHP, userMaxHP)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", healLogs))
		}

		if monster.Monster.Stats.Recovery > 0.0 {
			healLogs := ""
			monsterHeal := int(monster.Monster.Stats.HP * monster.Monster.Stats.Recovery)
			if monsterHeal+monsterHP > int(monster.Monster.Stats.HP) {
				monsterHP = int(monster.Monster.Stats.HP)
			} else {
				monsterHP += monsterHeal
			}
			healLogs += fmt.Sprintf("				__**%s**__ **HEALED** for %v HP.\n", monster.Name, monsterHeal)
			healLogs += fmt.Sprintf("				__**%s**__'s HP: %v/%v!\n", monster.Name, monsterHP, monsterMaxHp)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", healLogs))
		}

	}
	users[0].Buffs = nil
	levelCap, err := a.levels.ReadDocument("levelCap")
	if err != nil {
		a.log.Errorf("error retrieving current levelCap: %v", err)
		return nil, err
	}
	if battleWin {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- %s WON THE BATTLE.  GETTING RESULTS. ----------------------------**", users[0].User.Name))
		userClassInfo := *users[0].User.ClassMap[users[0].User.CurrentClass]
		expGainRate, err := a.GetExpGainRate("exp")
		if err != nil {
			return nil, err
		}
		monsterExp := int64(monster.Exp * float64(*expGainRate))
		a.log.Infof("monster Exp: %v", monster.Exp)
		a.log.Infof("monster Ely: %v", monster.Ely)
		if users[0].User.AscensionLevel > 0 && userClassInfo.Level == levelCap.Value {
			newExp := users[0].User.AscensionExp
			newExp += monsterExp
			users[0].User.AscensionExp = newExp
		} else {
			userClassInfo.Exp += monsterExp
		}
		monsterEly := int64(monster.Ely * float64(*expGainRate))
		*users[0].User.Ely += monsterEly
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%s*** Ely!", users[0].User.Name, utils.String(monsterExp), utils.String(monsterEly)))
		newUserClassInfo, newAdventureLog, newUser, err := a.processLevelUps(userClassInfo, adventureLog, users[0].User, levelCap.Value)
		adventureLog = newAdventureLog
		if err != nil {
			a.log.Errorf("error processing level ups: %v", err)
			return adventureLog, nil
		}
		item := a.getRandomItemDrop(*userWeapon, dropRange, *randGenerator, nil, 1, nil)
		if item != nil {
			if users[0].User.Inventory.Equipment == nil {
				users[0].User.Inventory.Equipment = make(map[string]int)
				users[0].User.Inventory.Event = make(map[string]int)
				users[0].User.Inventory.Consume = make(map[string]int)
			}
			if users[0].User.Inventory.Equipment[item.Name] == 0 && len(users[0].User.Inventory.Equipment) >= 45 {
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired nothing as their inventory is full!", users[0].User.Name))
			} else {
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired a **%s - Level %v %s**", users[0].User.Name, item.Name, *item.LevelRequirement, *item.Type.WeaponType))
				users[0].User.Inventory.Equipment[item.Name]++
			}
		}
		users[0].User.ClassMap[users[0].User.CurrentClass] = &newUserClassInfo
		users[0].User.AscensionExp = newUser.AscensionExp
		users[0].User.AscensionLevel = newUser.AscensionLevel
		users[0].User.LastActionTime = time.Now()
		//TODO: DISABLE WHEN RUNNING LOCAL
		_, err = a.users.UpdateDocument(users[0].User.ID, users[0].User)
	} else {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- %s LOST THE BATTLE. ----------------------------**", users[0].User.Name))
	}
	if err != nil {
		a.log.Errorf("failed to update user doc with error: %v", err)
		return adventureLog, nil
	}
	return adventureLog, nil
}

func (a *adventure) partyBattleLog(users []*models.UserBlob, encounteredMonsters []*models.MonsterBlob, adventureLog []string, primaryUser *models.User, dropRange models.LevelRange) ([]string, error) {
	battleWin := false
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	totalExpReward := int64(0)
	totalElyReward := int64(0)
	partyBonus := 1.0 + (float64(len(users)/10.0) * 3.5)
	monsterNames := ""
	for _, monster := range encounteredMonsters {
		totalExpReward += int64(float64(monster.Exp) * partyBonus)
		totalElyReward += int64(float64(monster.Ely) * partyBonus)
	}
	if len(encounteredMonsters) > 1 {
		for i, monster := range encounteredMonsters {
			if i == len(encounteredMonsters)-1 {
				monsterNames += "and a " + monster.Name
			} else {
				monsterNames += monster.Name + ", "
			}
		}
	} else {
		monsterNames += encounteredMonsters[0].Name
	}
	adventureLog = append(adventureLog, fmt.Sprintf("__**The Party**__ has encountered a __**%s**__", monsterNames))
	users, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.BATTLESTARTTRAIT, true)
	adventureLog = append(adventureLog, buffString)
combat:
	for a.checkGroupDeaths(users, encounteredMonsters) {
		//Party will target enemies in order of how they spawn.
		//Enemies will attack party members randomly.
		//battle continues until one side is no longer able to fight.
		userLogs := ""
		for i, user := range users {
			if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.SUMMONTRAIT && user.JobClass.Trait.ActivationRate != nil {
				summonChance := rand.Float64()

				if summonChance <= *user.JobClass.Trait.ActivationRate {
					summonLogs, newSummons := a.battle.GenerateSummons(user.User.Name, user.Summons, *user.JobClass.Trait.Summon, *user.BattleStats)
					adventureLog = append(adventureLog, summonLogs)
					user.Summons = newSummons
				}
			}
			userLog, damage, statusAilment := a.battle.DetermineHit(randGenerator, user.User.Name, encounteredMonsters[0].Name, *user.BattleStats, *encounteredMonsters[0].BattleStats, &user.Weapon, user.JobClass, &user.UserLevel, false, user.HitCount)
			userLogs += userLog + "\n"
			if len(user.Summons) > 0 {
				for i, summon := range user.Summons {
					summonLogs, summonDamage, _ := a.battle.DetermineHit(randGenerator, summon.Name+" "+strconv.Itoa(i+1), encounteredMonsters[0].Name, summon.StatModifier, *encounteredMonsters[0].BattleStats, &user.Weapon, user.JobClass, &user.UserLevel, false, user.HitCount)
					userLogs += summonLogs + "\n"
					damage += summonDamage
				}
			}
			currentMonsterHP := int(encounteredMonsters[0].CurrentHP)
			currentMonsterHP = ((int(currentMonsterHP) - int(damage)) + int(math.Abs(float64(currentMonsterHP-damage)))) / 2

			monsterMaxHp := int(encounteredMonsters[0].StatModifier.HP)
			if damage > 0 {
				userLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", encounteredMonsters[0].Name, currentMonsterHP, monsterMaxHp)
			}
			if statusAilment != nil {
				if encounteredMonsters[0].Debuffs == nil {
					encounteredMonsters[0].Debuffs = make(map[string]models.CrowdControlTrait)
				}
				monsterDebuffLogs, newMonster := a.battle.InflictStatusAilmentMonster(encounteredMonsters[0], *statusAilment)
				encounteredMonsters[0] = newMonster
				userLogs += monsterDebuffLogs
			}
			if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.AFTERATTACKTRAIT && user.JobClass.Trait.ActivationRate != nil {
				activationChance := rand.Float64()
				if user.JobClass.Trait.CrowdControl != nil && activationChance <= *user.JobClass.Trait.ActivationRate {
					userLogs += fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!\n", user.User.Name, user.JobClass.Trait.Name)
					if user.JobClass.Trait.CrowdControl.Type == "drain" {
						currentMonsterHP = int(float64(currentMonsterHP) * 0.975)
						drainHp := int(float64(users[i].MaxHP) * 0.2)
						if users[i].CurrentHP+drainHp > users[i].MaxHP {
							users[i].CurrentHP = users[i].MaxHP
						} else {
							users[i].CurrentHP += drainHp
						}
						userLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, drainHp)
						userLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, users[i].CurrentHP, user.MaxHP)

					}
				}
			}
			encounteredMonsters[0].CurrentHP = int32(currentMonsterHP)
			if encounteredMonsters[0].CurrentHP <= 0 {
				userLogs += fmt.Sprintf("__**%s**__ **has successfully defeated the** __**%s**__!\n", user.User.Name, encounteredMonsters[0].Name)
				copy(encounteredMonsters[0:], encounteredMonsters[0+1:]) // Shift a[i+1:] left one index.
				encounteredMonsters[len(encounteredMonsters)-1] = nil    // Erase last element (write zero value).
				encounteredMonsters = encounteredMonsters[:len(encounteredMonsters)-1]
			}
			if len(encounteredMonsters) == 0 {
				adventureLog = append(adventureLog, fmt.Sprintf("%s", userLogs))
				battleWin = true
				break combat
			}
		}
		buffedUsers, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.REACTIVETRAIT, false)
		users = buffedUsers
		userLogs += buffString
		var aliveMonsters []*models.MonsterBlob
		adventureLog = append(adventureLog, fmt.Sprintf("%s", userLogs))
		enemiesLog := ""
		for _, monster := range encounteredMonsters {
			aliveMonsters = append(aliveMonsters, monster)
		}
		monsterDied := false
		for i, monster := range encounteredMonsters {
			if monster.CurrentHP > 0 && monster.Debuffs != nil {
				for ailment, debuff := range monster.Debuffs {
					if ailment == "poison" || ailment == "bleed" || ailment == "burn" {
						damageOvertime := int32(0)
						switch ailment {
						case "poison":
							damageOvertime = int32(float64(monster.BattleStats.HP) * 0.05)
							break
						case "burn":
							damageOvertime = int32(float64(monster.BattleStats.HP) * 0.025)
							break
						case "bleed":
							damageOvertime = int32(float64(monster.CurrentHP) * 0.1)
							break
						}
						monster.CurrentHP = int32((monster.CurrentHP-damageOvertime)+int32(math.Abs(float64(monster.CurrentHP-damageOvertime)))) / 2
						enemyAilmentLog := ""
						enemyAilmentLog += fmt.Sprintf("				**%s lost %v HP!** due to **%s**.\n", monster.Name, int(damageOvertime), ailment)
						enemyAilmentLog += fmt.Sprintf("				__**%s**__'s HP: %v/%v", monster.Name, monster.CurrentHP, int(monster.StatModifier.HP))
						aliveMonsters[i] = monster
						if monster.CurrentHP <= 0 {
							enemyAilmentLog += fmt.Sprintf("\n__**The Party**__ **has successfully defeated the** __**%s**__!\n", monster.Name)
							monsterDied = true
						}
						adventureLog = append(adventureLog, fmt.Sprintf("%s", enemyAilmentLog))
					}
					if debuff.Bind {
						encounteredMonsters[i].Bound = true
					}

					monsterDebuffLogs, newMonster := a.battle.DecreaseMonsterDebuffDuration(monster)
					if newMonster != nil {
						monster = newMonster
					}
					if monsterDebuffLogs != "" {
						adventureLog = append(adventureLog, monsterDebuffLogs)
					}
				}

			}
		}
		if monsterDied {
			for _, monster := range aliveMonsters {
				encounteredMonsters = []*models.MonsterBlob{}
				if monster.CurrentHP > 0 {
					encounteredMonsters = append(encounteredMonsters, monster)
				}
			}
		}
		if len(encounteredMonsters) == 0 {
			battleWin = true
			break combat
		}
		for _, monster := range encounteredMonsters {
			if !monster.Bound {
				targetedUser := randGenerator.Intn(len(users))
				monsterLog, damage, _ := a.battle.DetermineHit(randGenerator, monster.Name, users[targetedUser].User.Name, *monster.StatModifier, *users[targetedUser].BattleStats, nil, users[targetedUser].JobClass, nil, false, 1)
				users[targetedUser].CurrentHP = ((users[targetedUser].CurrentHP - int(damage)) + int(math.Abs(float64(users[targetedUser].CurrentHP-damage)))) / 2
				enemiesLog += "				" + monsterLog + "\n"
				if damage > 0 {
					enemiesLog += fmt.Sprintf("				__**%s**__'s HP: %v/%v\n", users[targetedUser].User.Name, users[targetedUser].CurrentHP, users[targetedUser].MaxHP)
					if len(users[targetedUser].Summons) > 0 {
						var aliveSummons []models.Summons
						for i, summon := range users[targetedUser].Summons {
							if users[targetedUser].Summons[i].Duration == nil {
								users[targetedUser].Summons[i].StatModifier.HP -= float64(damage)
								enemiesLog += fmt.Sprintf("\n				 __**%s**__ has also taken **%v** damage!\n", summon.Name+" "+strconv.Itoa(i+1), int(damage))
								if users[targetedUser].Summons[i].StatModifier.HP > 0.0 {
									enemiesLog += fmt.Sprintf("				__**%s**__'s remaining HP: **%v**\n", summon.Name+" "+strconv.Itoa(i+1), int(users[targetedUser].Summons[i].StatModifier.HP))
									aliveSummons = append(aliveSummons, users[targetedUser].Summons[i])
								} else {
									enemiesLog += fmt.Sprintf("__**%s**__'s has died!\n", summon.Name+" "+strconv.Itoa(i+1))
								}
							} else {
								aliveSummons = append(aliveSummons, users[targetedUser].Summons[i])
							}
						}
						users[targetedUser].Summons = aliveSummons
					}
				}
				if users[targetedUser].CurrentHP <= 0 {
					if users[targetedUser].JobClass.Trait != nil && users[targetedUser].JobClass.Trait.Type == globals.DEATHTRAIT && *users[targetedUser].JobClass.Trait.UsageCount > 0 {
						tempTargetUser := users[targetedUser]
						buffedUsers, buffString := a.battle.DetermineTraitActivations([]*models.UserBlob{tempTargetUser}, adventureLog, globals.DEATHTRAIT, false)
						enemiesLog += buffString
						tempTargetUser = buffedUsers[0]
						tempTargetUser.MaxHP = int(tempTargetUser.BattleStats.HP)
						tempTargetUser.CurrentHP = tempTargetUser.MaxHP
						skillUsage := *tempTargetUser.JobClass.Trait.UsageCount
						skillUsage--
						tempTargetUser.JobClass.Trait.UsageCount = &skillUsage
						users[targetedUser] = tempTargetUser
					} else {
						enemiesLog += fmt.Sprintf("				**%s was killed by %s!**\n", users[targetedUser].User.Name, monster.Name)
						copy(users[targetedUser:], users[targetedUser+1:]) // Shift a[i+1:] left one index.
						users[len(users)-1] = nil                          // Erase last element (write zero value).
						users = users[:len(users)-1]
					}
				}
				a.log.Debugf("users: %v", users)
				if len(users) == 0 {
					adventureLog = append(adventureLog, fmt.Sprintf("%s", enemiesLog))
					break combat
				}
			} else {
				enemiesLog += fmt.Sprintf("				**%s** can't move due to being bound!\n", monster.Name)
			}
		}
		adventureLog = append(adventureLog, fmt.Sprintf("%s", enemiesLog))
		healLogs := ""
		for i, user := range users {
			buffLogs, changedUser := a.battle.DecreaseUserBuffDuration(users[i])
			if changedUser != nil {
				users[i] = changedUser
			}
			healLogs += buffLogs
			if len(users[i].Summons) > 0 {
				decreaseSummonLog, newSummons := a.battle.DecreaseSummonDuration(user.User.Name, users[i].Summons)
				healLogs += decreaseSummonLog
				users[i].Summons = newSummons
			}
			if user.JobClass.Trait != nil {
				buffedUser, buffString := a.battle.DetermineTraitActivations([]*models.UserBlob{user}, adventureLog, globals.REACTIVETRAIT, false)
				healLogs += buffString
				user = buffedUser[0]
				user.MaxHP = int(user.BattleStats.HP)
				users[i] = user
			}
			userHeal := int(user.BaseStats.HP * user.BattleStats.Recovery)
			if user.CurrentHP != int(user.MaxHP) && user.BattleStats.Recovery > 0.0 {
				if userHeal+user.CurrentHP > int(user.MaxHP) {
					user.CurrentHP = int(user.MaxHP)
					healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, userHeal)
					healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, user.CurrentHP, user.MaxHP)
				} else {
					user.CurrentHP += userHeal
					healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, userHeal)
					healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, user.CurrentHP, user.MaxHP)
				}
			}
			users[i].CurrentHP = user.CurrentHP

		}
		for _, user := range users {
			if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.HEALTRAIT {
				buffedUser, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.HEALTRAIT, false)
				healLogs += buffString
				users = buffedUser
			}
		}
		if healLogs != "" {
			adventureLog = append(adventureLog, fmt.Sprintf("%s", healLogs))
		}
		enemyHealLogs := ""
		for i, monster := range encounteredMonsters {
			if monster.StatModifier.Recovery > 0.0 {
				monsterHeal := int32(monster.StatModifier.HP * monster.StatModifier.Recovery)
				if monsterHeal+monster.CurrentHP > int32(monster.StatModifier.HP) {
					monster.CurrentHP = int32(monster.StatModifier.HP)
				} else {
					monster.CurrentHP += monsterHeal
				}
				encounteredMonsters[i].CurrentHP = monster.CurrentHP
				enemyHealLogs += fmt.Sprintf("				__**%s**__ **HEALED** for %v HP.\n", monster.Name, monsterHeal)
				enemyHealLogs += fmt.Sprintf("				__**%s**__'s HP: %v/%v!\n", monster.Name, monsterHeal+monster.CurrentHP, strconv.FormatFloat(monster.StatModifier.HP, 'f', -1, 64))
			}
		}
	}

	if battleWin {
		adventureLog = append(adventureLog, "**---------------------------- THE PARTY WON THE BATTLE.  GETTING RESULTS. ----------------------------**")
		levelCap, err := a.levels.ReadDocument("levelCap")
		if err != nil {
			a.log.Errorf("error retrieving current levelCap: %v", err)
			return nil, err
		}
		expGainRate, err := a.GetExpGainRate("exp")
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			user.Buffs = nil
			user.Debuffs = nil
			userInfo := user.User
			userClassInfo := *user.User.ClassMap[user.User.CurrentClass]
			monsterExp := totalExpReward / int64(len(users)) * int64(*expGainRate)
			if user.User.AscensionLevel > 0 && userClassInfo.Level == levelCap.Value {
				newExp := user.User.AscensionExp
				newExp += monsterExp
				user.User.AscensionExp = newExp
			} else {
				userClassInfo.Exp += monsterExp
			}
			monsterEly := totalElyReward / int64(len(users)) * int64(*expGainRate)
			oldEly := *user.User.Ely
			oldEly += totalElyReward / int64(len(users)) * int64(*expGainRate)
			userInfo.Ely = &oldEly
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", user.User.Name, utils.String(monsterExp), monsterEly))
			newUserClassInfo, newAdventureLog, newUser, err := a.processLevelUps(userClassInfo, adventureLog, user.User, levelCap.Value)
			if err != nil {
				a.log.Errorf("error processing level ups: %v", err)
				return adventureLog, nil
			}
			userInfo.ClassMap[user.User.CurrentClass] = &newUserClassInfo
			userInfo.AscensionLevel = newUser.AscensionLevel
			userInfo.AscensionExp = newUser.AscensionExp
			adventureLog = newAdventureLog
			if primaryUser.ID == userInfo.ID {
				userInfo.LastActionTime = time.Now()
			}
			item := a.getRandomItemDrop(user.Weapon, dropRange, *randGenerator, nil, len(users), nil)
			if item != nil {
				if userInfo.Inventory.Equipment == nil {
					userInfo.Inventory.Equipment = make(map[string]int)
					userInfo.Inventory.Event = make(map[string]int)
					userInfo.Inventory.Consume = make(map[string]int)
				}
				if userInfo.Inventory.Equipment[item.Name] == 0 && len(userInfo.Inventory.Equipment) >= 45 {
					adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired nothing as their inventory is full!", userInfo.Name))
				} else {
					adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired a **%s - Level %v %s**", userInfo.Name, item.Name, *item.LevelRequirement, *item.Type.WeaponType))
					userInfo.Inventory.Equipment[item.Name]++
				}
			}
			//TODO: DISABLE WHEN RUNNING LOCAL
			_, err = a.users.UpdateDocument(userInfo.ID, userInfo)
			if err != nil {
				a.log.Errorf("failed to update user doc with error: %v", err)
				return adventureLog, nil
			}
		}

	} else {
		adventureLog = append(adventureLog, "**---------------------------- THE PARTY LOST THE BATTLE. ----------------------------**")
	}
	a.log.Debugf("encounteredMonsters: %v", encounteredMonsters)
	return adventureLog, nil
}

func (a *adventure) bossBattleLog(users []*models.UserBlob, boss models.MonsterBlob, primaryUserId string) ([]string, error) {
	battleWin := false
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	bossMaxHp := int(boss.Monster.Stats.HP) * len(users)
	bossCurrentHp := bossMaxHp
	partyBonus := 1.0
	if len(users) > 1 {
		partyBonus = (float64(len(users)) / 10.0) + 1.0
	}
	bossHPPercentage := float64(bossCurrentHp) / float64(bossMaxHp)
	var adventureLog []string
	for _, skill := range *boss.Monster.Skills {
		curCd := int32(0)
		skill.CurrentCoolDown = &curCd
	}

	activeSkills := 1
	enraged := false
	adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BOSS ENCOUNTER------------------------**\n__**The Party**__ has encountered **%s**, __**%s**__.\n**------------------------BOSS ENCOUNTER------------------------**", boss.Name, *boss.Monster.BossTitle))
	users, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.BATTLESTARTTRAIT, true)
	adventureLog = append(adventureLog, buffString)
bossBattle:
	for len(users) != 0 && bossCurrentHp != 0 {
		bossLogs := ""
		bossBind := false
		for ailment, debuff := range boss.Debuffs {
			bossStatusAilmentLogs := ""
			if ailment == "poison" || ailment == "bleed" || ailment == "burn" {
				damageOvertime := 0
				switch ailment {
				case "poison":
					damageOvertime = int(float64(bossMaxHp) * 0.05)
					break
				case "burn":
					damageOvertime = int(float64(bossMaxHp) * 0.025)
					break
				case "bleed":
					damageOvertime = int(float64(bossCurrentHp) * 0.1)
					break
				}
				bossCurrentHp = int((bossCurrentHp-damageOvertime)+int(math.Abs(float64(bossCurrentHp-damageOvertime)))) / 2
				bossStatusAilmentLogs += fmt.Sprintf("**%s lost %v HP!** due to **%s**.\n", boss.Name, int(damageOvertime), ailment)
				bossHPPercentage = float64(bossCurrentHp) / float64(bossMaxHp)
				bossStatusAilmentLogs += fmt.Sprintf("__**%s**__'s **HP: %s%%/100%%**\n", boss.Name, fmt.Sprintf("%.2f", bossHPPercentage*100))
				if bossCurrentHp <= 0 {
					bossStatusAilmentLogs += fmt.Sprintf("__**The Party**__ **has successfully defeated ** __**%s**__!\n", boss.Name)
					adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN BOSS STATUS AILMENT TURN------------------------**\n%s**------------------------BEGIN BOSS STATUS AILMENT TURN------------------------**", bossStatusAilmentLogs))
					battleWin = true
					break bossBattle
				}
			}
			bossDebuffLogs, newBoss := a.battle.DecreaseMonsterDebuffDuration(&boss)
			if newBoss != nil {
				boss = *newBoss
			}
			if bossDebuffLogs != "" {
				bossStatusAilmentLogs += bossDebuffLogs
			}
			if debuff.Bind {
				a.log.Infof("ailment: %v", ailment)
				bossBind = true
			}
			if bossStatusAilmentLogs != "" {
				adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN BOSS STATUS AILMENT TURN------------------------**\n%s**------------------------BEGIN BOSS STATUS AILMENT TURN------------------------**", bossStatusAilmentLogs))
			}
		}
		userLogs := ""
		for i, user := range users {
			if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.SUMMONTRAIT && user.JobClass.Trait.ActivationRate != nil {
				summonChance := rand.Float64()

				if summonChance <= *user.JobClass.Trait.ActivationRate {
					summonLogs, newSummons := a.battle.GenerateSummons(user.User.Name, user.Summons, *user.JobClass.Trait.Summon, *user.BattleStats)
					userLogs += summonLogs + "\n"
					user.Summons = newSummons
				}
			}
			damage := 0
			var statusAilment *models.CrowdControlTrait

			bound := false
			for ailment, debuff := range user.Debuffs {
				if debuff.Bind {
					userLogs += fmt.Sprintf("__**%s**__ is currently under the effects of __**%s**__ for **%v turn(s)**.\n", user.User.Name, ailment, debuff.CrowdControlTime)
					bound = true
				}
			}
			if !bound {
				userLog, userDamage, userInflictedAilment := a.battle.DetermineHit(randGenerator, user.User.Name, boss.Name, *user.BattleStats, *boss.BattleStats, &user.Weapon, user.JobClass, &user.UserLevel, true, user.HitCount)
				userLogs += userLog + "\n"
				damage += userDamage
				statusAilment = userInflictedAilment
			}
			if len(user.Summons) > 0 {
				for i, summon := range user.Summons {
					summonLog, summonDamage, _ := a.battle.DetermineHit(randGenerator, summon.Name+" "+strconv.Itoa(i+1), boss.Name, summon.StatModifier, *boss.BattleStats, &user.Weapon, user.JobClass, &user.UserLevel, true, user.HitCount)
					userLogs += summonLog + "\n"
					damage += summonDamage
				}
			}
			bossCurrentHp = ((int(bossCurrentHp) - int(damage)) + int(math.Abs(float64(bossCurrentHp-damage)))) / 2
			bossHPPercentage = float64(bossCurrentHp) / float64(bossMaxHp)
			if damage > 0 {
				userLogs += fmt.Sprintf("__**%s**__'s **HP: %s%%/100%%**\n", boss.Name, fmt.Sprintf("%.2f", bossHPPercentage*100))
				if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.AFTERATTACKTRAIT && user.JobClass.Trait.ActivationRate != nil {
					activationChance := rand.Float64()
					if user.JobClass.Trait.CrowdControl != nil && activationChance <= *user.JobClass.Trait.ActivationRate {
						userLogs += fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!\n", user.User.Name, user.JobClass.Trait.Name)
						if user.JobClass.Trait.CrowdControl.Type == "drain" {
							bossCurrentHp = int(float64(bossCurrentHp) - float64(bossCurrentHp)*0.05)
							drainHp := int(float64(users[i].MaxHP) * 0.10)
							if users[i].CurrentHP+drainHp > users[i].MaxHP {
								users[i].CurrentHP = users[i].MaxHP
							} else {
								users[i].CurrentHP += drainHp
							}
							bossHPPercentage = float64(bossCurrentHp) / float64(bossMaxHp)
							userLogs += fmt.Sprintf("__**%s**__'s **HP: %s%%/100%%**\n", boss.Name, fmt.Sprintf("%.2f", bossHPPercentage*100))
							userLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, drainHp)
							userLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, users[i].CurrentHP, user.MaxHP)
						}
					}
				}
			}

			if statusAilment != nil {
				debuffLogs, newBoss := a.battle.InflictStatusAilmentMonster(&boss, *statusAilment)
				boss = *newBoss
				userLogs += debuffLogs
				if statusAilment.Bind {
					bossBind = true
				}
			}

			if bossCurrentHp <= 0 {
				userLogs += fmt.Sprintf("__**The Party**__ **has successfully defeated ** __**%s**__!\n", boss.Name)
				adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN PARTY ATTACK TURN------------------------**\n%s**------------------------END PARTY ATTACK TURN------------------------**", userLogs))
				battleWin = true
				break bossBattle
			}
		}
		buffedUsers, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.REACTIVETRAIT, false)
		users = buffedUsers
		userLogs += buffString
		adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN PARTY ATTACK TURN------------------------**\n%s**------------------------END PARTY ATTACK TURN------------------------**", userLogs))
		phase, newActiveSkills := a.checkPhaseStatus(bossHPPercentage, &boss.Monster, activeSkills)
		if phase != "" && newActiveSkills != activeSkills {
			activeSkills = newActiveSkills
			enragedText := ""
			if activeSkills == 4 {
				enragedText += fmt.Sprintf("__**%s has become enraged!**__\n", boss.Name)
				enraged = true
			}
			adventureLog = append(adventureLog, fmt.Sprintf("**------------------------PHASE %v------------------------**\n%s\n%s**------------------------PHASE %v------------------------**", activeSkills, phase, enragedText, activeSkills))
		}
		active := randGenerator.Float64()

		if !bossBind && (active > *boss.Monster.IdleTime || enraged) {
			skills := *boss.Monster.Skills
			var availableSkills []*models.BossSkill
			for i := 0; i < activeSkills; i++ {
				if *skills[i].CurrentCoolDown == 0 {
					availableSkills = append(availableSkills, skills[i])
				}
			}
			var skill *models.BossSkill
			if len(availableSkills) != 0 {
				skillGen := randGenerator.Intn(len(availableSkills))
				skill = availableSkills[skillGen]
				for _, usedSkill := range skills {
					if usedSkill.Name == skill.Name {
						usedSkill.CurrentCoolDown = &usedSkill.CoolDown
					}
				}
			} else {
				skill = nil
			}
			var alivePlayers []*models.UserBlob
			for _, user := range users {
				alivePlayers = append(alivePlayers, user)
			}
			if skill != nil && skill.AoE {
				for i, user := range users {
					updatedUser, bossDamageLog, damage := a.battle.DetermineBossDamage(randGenerator, *user, &boss.Monster, skill)
					updatedUser.CurrentHP = ((updatedUser.CurrentHP - int(damage)) + int(math.Abs(float64(updatedUser.CurrentHP-damage)))) / 2
					bossLogs += bossDamageLog + "\n"
					if damage > 0 {
						bossLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", updatedUser.User.Name, updatedUser.CurrentHP, updatedUser.MaxHP)
					}

					if updatedUser.CurrentHP <= 0 {
						if updatedUser.JobClass.Trait != nil && updatedUser.JobClass.Trait.Type == globals.DEATHTRAIT && *updatedUser.JobClass.Trait.UsageCount > 0 {
							buffedUsers, buffString := a.battle.DetermineTraitActivations([]*models.UserBlob{updatedUser}, adventureLog, globals.DEATHTRAIT, false)
							updatedUser = buffedUsers[0]
							bossLogs += buffString
							updatedUser.MaxHP = int(updatedUser.BattleStats.HP)
							updatedUser.CurrentHP = updatedUser.MaxHP
							skillUsage := *updatedUser.JobClass.Trait.UsageCount
							skillUsage--
							updatedUser.JobClass.Trait.UsageCount = &skillUsage
						} else {
							bossLogs += fmt.Sprintf("**%s was killed by %s!**\n", user.User.Name, boss.Name)
						}
					}
					alivePlayers[i] = updatedUser
				}
				users = []*models.UserBlob{}
				for _, user := range alivePlayers {
					if user.CurrentHP > 0 {
						users = append(users, user)
					}
				}
			} else {
				targetedUser := randGenerator.Intn(len(users))
				updatedUser, bossDamageLog, damage := a.battle.DetermineBossDamage(randGenerator, *users[targetedUser], &boss.Monster, skill)
				updatedUser.CurrentHP = ((updatedUser.CurrentHP - int(damage)) + int(math.Abs(float64(updatedUser.CurrentHP-damage)))) / 2
				bossLogs += bossDamageLog + "\n"
				if damage > 0 {
					bossLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", updatedUser.User.Name, updatedUser.CurrentHP, updatedUser.MaxHP)
				}
				if updatedUser.CurrentHP <= 0 {
					if updatedUser.JobClass.Trait != nil && updatedUser.JobClass.Trait.Type == globals.DEATHTRAIT && *updatedUser.JobClass.Trait.UsageCount > 0 {
						buffedUsers, buffString := a.battle.DetermineTraitActivations([]*models.UserBlob{updatedUser}, adventureLog, globals.DEATHTRAIT, false)
						updatedUser = buffedUsers[0]
						bossLogs += buffString
						updatedUser.MaxHP = int(updatedUser.BattleStats.HP)
						updatedUser.CurrentHP = updatedUser.MaxHP
						skillUsage := *updatedUser.JobClass.Trait.UsageCount
						skillUsage--
						updatedUser.JobClass.Trait.UsageCount = &skillUsage
						users[targetedUser] = updatedUser
					} else {
						bossLogs += fmt.Sprintf("**%s was killed by %s!**\n", updatedUser.User.Name, boss.Name)
						copy(users[targetedUser:], users[targetedUser+1:]) // Shift a[i+1:] left one index.
						users[len(users)-1] = nil                          // Erase last element (write zero value).
						users = users[:len(users)-1]
					}
				} else {
					users[targetedUser] = updatedUser
				}
			}

			if len(users) == 0 {
				bossLogs += fmt.Sprintf("__**The Party**__ was completely wiped out by __**%s**__ .\n", boss.Name)
				adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN BOSS ATTACK TURN------------------------**\n%s**------------------------END BOSS ATTACK TURN------------------------**", bossLogs))
				break bossBattle
			}
			for _, skill := range skills {
				if *skill.CurrentCoolDown > 0 {
					curCoolDown := *skill.CurrentCoolDown
					curCoolDown--
					skill.CurrentCoolDown = &curCoolDown
				}
			}
			boss.Monster.Skills = &skills
		} else {
			bossLogs += *boss.Monster.IdlePhrase + "\n"
		}
		adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN BOSS ATTACK TURN------------------------**\n%s**------------------------END BOSS ATTACK TURN------------------------**", bossLogs))

		healLogs := ""
		var alivePlayers []*models.UserBlob
		for _, user := range users {
			alivePlayers = append(alivePlayers, user)
		}
		playerDied := false
		for i, user := range users {
			for ailment, debuff := range user.Debuffs {
				if user.CurrentHP > 0 && debuff.CrowdControlTime != int32(0) && (debuff.Type == "poison" || debuff.Type == "bleed" || debuff.Type == "burn") {
					damageOvertime := 0
					if debuff.Type == "bleed" {
						damageOvertime = int(float64(user.CurrentHP) * 0.10)
					} else if debuff.Type == "poison" || debuff.Type == "burn" {
						damageOvertime = int(float64(user.MaxHP) * 0.075)
					}
					user.CurrentHP = ((user.CurrentHP - int(damageOvertime)) + int(math.Abs(float64(user.CurrentHP-damageOvertime)))) / 2
					healLogs += fmt.Sprintf("**%s lost %v HP!** due to **%s**. %s has the status ailment of **%s** for **%v turn(s)**.\n", user.User.Name, damageOvertime, ailment, user.User.Name, ailment, debuff.CrowdControlTime)
					healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", user.User.Name, user.CurrentHP, user.MaxHP)
					alivePlayers[i] = user
					if user.CurrentHP <= 0 {
						if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.DEATHTRAIT && *user.JobClass.Trait.UsageCount > 0 {
							buffedUsers, buffString := a.battle.DetermineTraitActivations([]*models.UserBlob{user}, adventureLog, globals.DEATHTRAIT, false)
							user = buffedUsers[0]
							healLogs += buffString
							user.MaxHP = int(user.BattleStats.HP)
							user.CurrentHP = user.MaxHP
							skillUsage := *user.JobClass.Trait.UsageCount
							skillUsage--
							user.JobClass.Trait.UsageCount = &skillUsage
						} else {
							healLogs += fmt.Sprintf("**%s was killed by %s!**\n", user.User.Name, boss.Name)
							playerDied = true
						}
						alivePlayers[i] = user
					}
				}
			}
			buffLogs, changedUser := a.battle.DecreaseUserDebuffDuration(users[i])
			if changedUser != nil {
				users[i].BattleStats = changedUser.BattleStats
			}
			if buffLogs != "" {
				healLogs += buffLogs
			}
		}
		if playerDied {
			for _, user := range alivePlayers {
				users = []*models.UserBlob{}
				if user.CurrentHP > 0 {
					users = append(users, user)
				}
			}
		}
		for i, user := range users {
			buffLogs, changedUser := a.battle.DecreaseUserBuffDuration(users[i])
			if changedUser != nil {
				users[i] = changedUser
			}
			healLogs += buffLogs
			if users[i].Summons != nil && len(users[i].Summons) > 0 {
				decreaseSummonLog, newSummons := a.battle.DecreaseSummonDuration(user.User.Name, users[i].Summons)
				healLogs += decreaseSummonLog
				users[i].Summons = newSummons
			}
			if user.CurrentHP != int(user.MaxHP) && user.BattleStats.Recovery > 0.0 {
				userHeal := int(user.BaseStats.HP * user.BattleStats.Recovery)
				if userHeal+user.CurrentHP > int(user.MaxHP) {
					user.CurrentHP = int(user.MaxHP)
					healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, userHeal)
					healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, user.CurrentHP, user.MaxHP)
				} else {
					user.CurrentHP += userHeal
					healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.User.Name, userHeal)
					healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.User.Name, user.CurrentHP, user.MaxHP)
				}
			}
			users[i].CurrentHP = user.CurrentHP
		}
		for _, user := range users {
			if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.HEALTRAIT {
				buffedUser, buffString := a.battle.DetermineTraitActivations(users, adventureLog, globals.HEALTRAIT, false)
				healLogs += buffString
				users = buffedUser
			}
		}
		if healLogs != "" {
			adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN PARTY HEAL TURN------------------------**\n%s**------------------------END PARTY HEAL TURN------------------------**", healLogs))
		}
	}

	if battleWin {
		adventureLog = append(adventureLog, "**---------------------------- PARTY WON THE BATTLE.  GETTING RESULTS. ----------------------------**")
		levelCap, err := a.levels.ReadDocument("levelCap")
		if err != nil {
			a.log.Errorf("error retrieving current levelCap: %v", err)
			return nil, err
		}
		expGainRate, err := a.GetExpGainRate("exp")
		if err != nil {
			return nil, err
		}
		for _, winningUsers := range users {
			winningUsers.Buffs = nil
			winningUsers.Debuffs = nil
			userInfo := winningUsers.User
			userClassInfo := *winningUsers.User.ClassMap[winningUsers.User.CurrentClass]
			userClassInfo, adventureLog = a.determineBossBonusDrop(winningUsers.User.Name, userClassInfo, boss.Monster, adventureLog)
			bossExp := int64(float64(boss.Exp*float64(*expGainRate)) * partyBonus)
			userClassInfo.Exp += bossExp
			if userInfo.AscensionLevel > 0 && userClassInfo.Level == levelCap.Value {
				newExp := userInfo.AscensionExp
				newExp += bossExp
				userInfo.AscensionExp = newExp
			} else {
				userClassInfo.Exp += bossExp
			}
			oldEly := *winningUsers.User.Ely
			bossEly := int64(float64(boss.Ely*float64(*expGainRate)) * partyBonus)
			oldEly += bossEly
			userInfo.ClassMap[userInfo.CurrentClass] = &userClassInfo
			userInfo.Ely = &oldEly
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", winningUsers.User.Name, utils.String(bossExp), bossEly))
			newUserClassInfo, newAdventureLog, newUser, err := a.processLevelUps(userClassInfo, adventureLog, winningUsers.User, levelCap.Value)
			if err != nil {
				a.log.Errorf("error processing level ups: %v", err)
				return adventureLog, nil
			}
			userInfo.ClassMap[userInfo.CurrentClass] = &newUserClassInfo
			userInfo.AscensionExp = newUser.AscensionExp
			userInfo.AscensionLevel = newUser.AscensionLevel
			adventureLog = newAdventureLog

			userInfo.LastBossActionTime = time.Now()
			item := a.getRandomItemDrop(*userInfo.ClassMap[userInfo.CurrentClass].Equipment.Weapon.Type.WeaponType, *boss.Monster.DropRange, *randGenerator, &boss.Name, len(users), &userInfo.CurrentClass)
			if item != nil {
				if userInfo.Inventory.Equipment == nil {
					userInfo.Inventory.Equipment = make(map[string]int)
					userInfo.Inventory.Event = make(map[string]int)
					userInfo.Inventory.Consume = make(map[string]int)
				}
				if userInfo.Inventory.Equipment[item.Name] == 0 && len(userInfo.Inventory.Equipment) >= 45 {
					adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired nothing as their inventory is full!", userInfo.Name))
				} else {
					adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ acquired a **%s - Level %v %s**", userInfo.Name, item.Name, *item.LevelRequirement, *item.Type.WeaponType))
					userInfo.Inventory.Equipment[item.Name]++
				}

			}
			//TODO: DISABLE WHEN RUNNING LOCAL
			_, err = a.users.UpdateDocument(userInfo.ID, userInfo)
			if err != nil {
				a.log.Errorf("failed to update winningUsers doc with error: %v", err)
				return adventureLog, nil
			}
		}

	} else {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- THE PARTY LOST THE BATTLE AGAINST** __**%s**__.**----------------------------**", boss.Name))
	}
	return adventureLog, nil
}

func (a *adventure) checkPhaseStatus(bossPercent float64, boss *models.Monster, phaseCount int) (string, int) {
	phases := *boss.Phases
	if bossPercent <= 0.15 && phaseCount < 4 {
		return phases[2], 4
	}
	if bossPercent <= 0.50 && phaseCount < 3 {
		return phases[1], 3
	}
	if bossPercent <= 0.75 && phaseCount < 2 {
		return phases[0], 2
	}
	return "", phaseCount
}

/*
	Applies buffs to players in a party and returns that party.
*/

func (a *adventure) checkGroupDeaths(users []*models.UserBlob, encounteredMonsters []*models.MonsterBlob) bool {
	playersAlive := len(users)
	for _, user := range users {
		if user.CurrentHP == 0 {
			playersAlive--
		}
		if playersAlive == 0 {
			return false
		}
	}
	monstersAlive := len(encounteredMonsters)
	for _, monster := range encounteredMonsters {
		if monster.CurrentHP == 0 {
			monstersAlive--
		}
		if monstersAlive == 0 {
			return false
		}
	}
	return true
}

func (a *adventure) processLevelUps(userClassInfo models.ClassInfo, adventureLog []string, user *models.User, levelCap int32) (models.ClassInfo, []string, *models.User, error) {
	if user.ClassMap[user.CurrentClass].Level < int32(200) {
		a.log.Infof("ascension is in fact 0")
		level, err := a.levels.ReadDocument(utils.ThirtyTwoBitIntToString(userClassInfo.Level))
		if err != nil {
			a.log.Errorf("error getting level data: %v", err)
			return userClassInfo, adventureLog, user, err
		}
		a.log.Infof("userclassinfo: exp: %v", userClassInfo.Exp)
		a.log.Infof("level: exp: %v", level.Exp)
		if levelCap == userClassInfo.Level {
			if user.AscensionLevel == int32(0) {
				userClassInfo.Exp = 0.0
				user.AscensionLevel = 1
				user.AscensionExp = 0
			}
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ ** has reached the level cap and unlocked Ascension Levels!  Ascension Levels will slowly boost stats and max out at Ascension Level 9999.  Good luck.**", user.Name))
			return userClassInfo, adventureLog, user, nil
		}
		if userClassInfo.Exp >= level.Exp {
			a.log.Infof("userclassinfo: ascension: %v", user.AscensionLevel)

			if userClassInfo.Level < levelCap {
				userClassInfo.Exp -= level.Exp
				userClassInfo.Level++
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ **LEVELED UP**!  Current Level: %v", user.Name, userClassInfo.Level))
			}
			if userClassInfo.Level == 50 {

				advanceJobs, err := a.classes.QueryDocuments(&[]models.QueryArg{{Path: "classRequirement", Op: "==", Value: user.CurrentClass}})
				if err != nil {
					a.log.Errorf("error querying for 2nd tier classes")
					return userClassInfo, adventureLog, user, err
				}
				possibleJobs := *advanceJobs
				jobText := ""
				for i, job := range possibleJobs {
					if i == 0 {
						jobText += job.Name
					} else {
						jobText = "either " + jobText + " or " + job.Name
					}
				}
				if advanceJobs != nil && len(*advanceJobs) > 0 {
					jobs := *advanceJobs
					adventureLog = append(adventureLog, fmt.Sprintf("Congratulations!  Now that you've reached level %v, you may use the **-classAdvance <Class> <Weapon>** command to advance to ***%s***", jobs[0].LevelRequirement, jobText))
				}
			} else if userClassInfo.Level == 100 {

				advanceJobs, err := a.classes.QueryDocuments(&[]models.QueryArg{{Path: "classRequirement", Op: "==", Value: user.CurrentClass}})
				if err != nil {
					a.log.Errorf("error querying for 3rd tier classes")
					return userClassInfo, adventureLog, user, err
				}
				possibleJobs := *advanceJobs
				jobText := ""
				for i, job := range possibleJobs {
					if i == 0 {
						jobText += job.Name
					} else {
						jobText = "either " + jobText + " or " + job.Name
					}
				}
				if advanceJobs != nil && len(*advanceJobs) > 0 {
					jobs := *advanceJobs
					adventureLog = append(adventureLog, fmt.Sprintf("Congratulations!  Now that you've reached level %v, you may use the **-classAdvance <Class> <Weapon>** command to advance to ***%s***", jobs[0].LevelRequirement, jobText))
				}
			} else if userClassInfo.Level == 150 {

				advanceJobs, err := a.classes.QueryDocuments(&[]models.QueryArg{{Path: "classRequirement", Op: "==", Value: user.CurrentClass}})
				if err != nil {
					a.log.Errorf("error querying for 4th tier classes")
					return userClassInfo, adventureLog, user, err
				}
				possibleJobs := *advanceJobs
				jobText := ""
				for i, job := range possibleJobs {
					if i == 0 {
						jobText += job.Name
					} else {
						jobText = "either " + jobText + " or " + job.Name
					}
				}
				if advanceJobs != nil && len(*advanceJobs) > 0 {
					jobs := *advanceJobs
					adventureLog = append(adventureLog, fmt.Sprintf("Congratulations!  Now that you've reached level %v, you may use the **-classAdvance <Class> <Weapon>** command to advance to ***%s***", jobs[0].LevelRequirement, jobText))
				}
			}
			a.log.Infof("userclassinfo: ascension: %v", user.AscensionLevel)
			return a.processLevelUps(userClassInfo, adventureLog, user, levelCap)
		} else {
			adventureLog = append(adventureLog, fmt.Sprintf("Current Exp: **%s/%s**", utils.String(userClassInfo.Exp), utils.String(level.Exp)))
		}

	} else if user.AscensionLevel > 0 && user.ClassMap[user.CurrentClass].Level == int32(200) {
		ascensionLevel, err := a.ascension.ReadDocument(utils.ThirtyTwoBitIntToString(user.AscensionLevel))
		if err != nil {
			a.log.Errorf("error getting ascension level data: %v", err)
			return userClassInfo, adventureLog, user, err
		}
		if user.AscensionExp >= ascensionLevel.Exp {
			user.AscensionExp -= ascensionLevel.Exp
			user.AscensionLevel++
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ **ASCENSION LEVEL UP**!  Current Ascension Level: %v", user.Name, user.AscensionLevel))
			return a.processLevelUps(userClassInfo, adventureLog, user, levelCap)
		} else if user.AscensionLevel == 9999 {
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ ** has reached the Ascension Level cap!  Congratulations!!!!**", user.Name))
			return userClassInfo, adventureLog, user, nil
		} else {
			adventureLog = append(adventureLog, fmt.Sprintf("Current Ascension Exp: **%s/%s**", utils.String(user.AscensionExp), utils.String(ascensionLevel.Exp)))
		}
	}
	return userClassInfo, adventureLog, user, nil
}

func (a *adventure) getRandomItemDrop(currentWeapon string, dropRange models.LevelRange, rand rand.Rand, boss *string, partySize int, currentClass *string) *models.Item {
	dropChance := rand.Float64()
	if partySize > 1 {
		dropChance = dropChance - (0.05 * float64(partySize))
	}
	if boss != nil && dropChance <= 0.80 {
		items, _ := a.item.QueryDocuments(&[]models.QueryArg{
			{
				Path:  "boss",
				Op:    "==",
				Value: *boss,
			},
		})
		bossDropChance := rand.Float64()
		if len(items) > 0 {
			item := rand.Intn(len(items))
			if bossDropChance <= 0.1 {
				for _, equip := range items {
					if *equip.Type.WeaponType == currentWeapon {
						return &equip
					}
				}
			}
			if bossDropChance <= 0.5 {
				var classItems []models.Item
				for _, equip := range items {
					if equip.RequiredClasses != nil && len(*equip.RequiredClasses) > 0 {
						for _, requiredClass := range *equip.RequiredClasses {
							if *requiredClass == *currentClass {
								classItems = append(classItems, equip)
							}
						}
					}
				}
				if classItems != nil && len(classItems) > 0 {
					item = rand.Intn(len(classItems))
					return &classItems[item]
				}
			}
			if bossDropChance <= 0.9 {
				var normalBossItems []models.Item
				for _, equip := range items {
					if equip.RequiredClasses == nil && equip.Type.Type == "armor" {
						normalBossItems = append(normalBossItems, equip)
					}
				}
				if normalBossItems != nil && len(normalBossItems) > 0 {
					item = rand.Intn(len(normalBossItems))
					return &normalBossItems[item]
				}
			}
			return &items[item]
		}
	}
	if dropChance <= 0.20 {
		items, err := a.item.QueryDocuments(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "<=",
				Value: dropRange.Max,
			},
			{
				Path:  "levelRequirement",
				Op:    ">=",
				Value: dropRange.Min,
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: currentWeapon,
			},
		})
		if err != nil {
			panic("failed to get items for drops")
		}
		if len(items) == 0 {
			return nil
		}
		//TODO: REPLACE THIS BY BATCHING OVER ITEMS TO ADD BOSS FIELD
		var droppableItems []models.Item
		for _, item := range items {
			if item.Boss == nil {
				droppableItems = append(droppableItems, item)
			}
		}
		if len(droppableItems) == 0 {
			return nil
		}
		item := rand.Intn(len(droppableItems))
		return &droppableItems[item]
	}
	if dropChance <= 0.90 {
		items, err := a.item.QueryDocuments(&[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "<=",
				Value: dropRange.Max,
			},
			{
				Path:  "levelRequirement",
				Op:    ">=",
				Value: dropRange.Min,
			},
			{
				Path:  "type.itemType",
				Op:    "==",
				Value: "armor",
			},
		})
		if len(items) == 0 {
			return nil
		}
		if err != nil {
			panic("failed to get items for drops")
		}
		//TODO: REPLACE THIS BY BATCHING OVER ITEMS TO ADD BOSS FIELD
		var droppableItems []models.Item
		for _, item := range items {
			if item.Boss == nil {
				droppableItems = append(droppableItems, item)
			}
		}
		if len(droppableItems) == 0 {
			return nil
		}
		item := rand.Intn(len(droppableItems))
		return &droppableItems[item]
	}
	return nil
}

func (a *adventure) addNewEquipmentSheet(equipSheet map[string]*models.OldEquipmentSheet, equipment *models.OldEquipmentSheet) map[string]*models.OldEquipmentSheet {
	if equipSheet[equipment.ID] == nil {
		equipSheet[equipment.ID] = equipment
	}
	return equipSheet
}

func (a *adventure) calculateBaseStat(user models.User, class models.StatModifier) (*models.StatModifier, error) {
	level := float64(user.ClassMap[user.CurrentClass].Level)
	levelModifier := float64((level / 100) + 1)
	bossMaxDps := 0.0
	bossMinDps := 0.0
	bossDefense := 0.0
	bossHp := 0.0
	bossRecv := 0.0
	bossCritDmg := 0.0
	bossCritRate := 0.0
	bossSkillProc := 0.0
	bossEvasion := 0.0
	bossAccuracy := 0.0
	bossSkillDmg := 0.0
	bossTdd := 0.0
	bossDamageMit := 0.0
	if user.ClassMap[user.CurrentClass].BossBonuses != nil && len(user.ClassMap[user.CurrentClass].BossBonuses) > 0 {
		for _, bonus := range user.ClassMap[user.CurrentClass].BossBonuses {
			bossMaxDps += bonus.MaxDPS
			bossMinDps += bonus.MinDPS
			bossDefense += bonus.Defense
			bossHp += bonus.HP
			bossRecv += bonus.Recovery
			bossCritDmg += bonus.CriticalDamageModifier
			bossCritRate += bonus.CriticalRate
			bossSkillProc += bonus.SkillProcRate
			bossEvasion += bonus.Evasion
			bossAccuracy += bonus.Accuracy
			bossSkillDmg += bonus.SkillDamageModifier
			bossTdd += bonus.TargetDefenseDecrease
			bossDamageMit += bonus.DamageMitigation
		}
	}
	baseStats := models.StatModifier{
		MaxDPS:                 getDynamicStat(20, levelModifier, class.MaxDPS) + bossMaxDps,
		MinDPS:                 getDynamicStat(20, levelModifier, class.MinDPS) + bossMinDps,
		Defense:                getDynamicStat(15, levelModifier, class.Defense) + bossDefense,
		HP:                     getDynamicStat(100, levelModifier, class.HP) + bossHp,
		Recovery:               getStaticStat(0.05, levelModifier, class.Recovery) + bossRecv,
		CriticalDamageModifier: getStaticStat(1.5, levelModifier, class.CriticalDamageModifier) + bossCritDmg,
		CriticalRate:           getStaticStat(0.05, levelModifier, class.CriticalRate) + bossCritRate,
		SkillProcRate:          getStaticStat(0.25, levelModifier, class.SkillProcRate) + bossSkillProc,
		Evasion:                getStaticStat(0.05, levelModifier, class.Evasion) + bossEvasion,
		Accuracy:               0.85*class.Accuracy + bossAccuracy,
		TargetDefenseDecrease:  a.getDiminishingStat(0, class.TargetDefenseDecrease),
		SkillDamageModifier:    class.SkillDamageModifier + bossSkillDmg,
		DamageMitigation:       class.DamageMitigation + bossDamageMit,
	}

	if user.AscensionLevel > 0 {
		ascensionLevel := float64(user.AscensionLevel)
		ascensionLevel = 800
		ascensionStats := models.StatModifier{
			MaxDPS:                 500 * ascensionLevel,
			MinDPS:                 500 * ascensionLevel,
			Defense:                10 * ascensionLevel,
			HP:                     500 * ascensionLevel,
			Recovery:               0,
			CriticalDamageModifier: 0,
			CriticalRate:           0.00001 * ascensionLevel,
			SkillProcRate:          0,
			Evasion:                0,
			Accuracy:               0.000020002 * ascensionLevel,
			TargetDefenseDecrease:  0.00002 * ascensionLevel,
			SkillDamageModifier:    0,
			DamageMitigation:       0.00001 * ascensionLevel,
		}
		baseStats.AddStatModifier(ascensionStats)
	}
	equip := user.ClassMap[user.CurrentClass].Equipment
	gearStats, err := a.getStatsFromGear(&equip)
	if err != nil {
		a.log.Errorf("error getting stats from gear: %v", err)
		return nil, err
	}
	baseStats.TargetDefenseDecrease = a.getDiminishingStat(baseStats.TargetDefenseDecrease, bossTdd)
	baseStats.AddStatModifier(*gearStats)
	if user.ClassMap[user.CurrentClass].SetBonuses != nil && len(user.ClassMap[user.CurrentClass].SetBonuses) > 0 {
		for _, bonus := range user.ClassMap[user.CurrentClass].SetBonuses {
			if bonus.CurrentlyEquipped >= bonus.RequiredPieces {
				baseStats.AddStatModifier(*bonus.Bonus)
			}
		}
	}
	return &baseStats, nil
}

func (a *adventure) getDiminishingStat(statTotal, addedStat float64) float64 {
	diminishedStat := 1.0 - statTotal
	diminishedStat *= addedStat
	return statTotal + diminishedStat
}

func (a *adventure) getStatsFromGear(equips *models.Equipment) (*models.StatModifier, error) {
	totalEquipStats := models.StatModifier{}
	totalEquipStats.AddStatModifier(*equips.Weapon.Stats)
	totalEquipStats.AddStatModifier(*equips.Top.Stats)
	totalEquipStats.AddStatModifier(*equips.Headpiece.Stats)
	totalEquipStats.AddStatModifier(*equips.Bottom.Stats)
	totalEquipStats.AddStatModifier(*equips.Glove.Stats)
	totalEquipStats.AddStatModifier(*equips.Shoes.Stats)
	if equips.Bindi != nil {
		totalEquipStats.AddStatModifier(*equips.Bindi.Stats)
		totalEquipStats.AddStatModifier(*equips.Glasses.Stats)
		totalEquipStats.AddStatModifier(*equips.Earring.Stats)
		totalEquipStats.AddStatModifier(*equips.Ring.Stats)
		totalEquipStats.AddStatModifier(*equips.Cloak.Stats)
		totalEquipStats.AddStatModifier(*equips.Stockings.Stats)
	}
	return &totalEquipStats, nil
}

func getDynamicStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * statModifier * math.Pow(levelModifier, 7)
}

func getStaticStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * levelModifier * statModifier
}
