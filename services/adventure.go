package services

import (
	"encoding/json"
	"fmt"
	"github.com/juju/loggo"
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
	GetBossBonus(id int32) (*models.BossBonus, error)
	ClassChange(id, class string, weapon *string) (*string, error)
	CreateParty(id string) (*string, error)
	JoinParty(partyId, id string) (*string, error)
	LeaveParty(id string) (*string, error)
	UpdateEquipmentPiece(id, equipment string) (*string, error)
	GetEquipmentPieceCost(id, equipment string) (*string, error)
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
}

type adventure struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	users     repositories.UserRepository
	levels    repositories.LevelRepository
	equipment repositories.EquipmentRepository
	config    repositories.ConfigRepository
	party     repositories.PartyRepository
	boss      repositories.BossRepository
	damage    Damage
	log       loggo.Logger
}

func NewAdventureService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, equips repositories.EquipmentRepository, levels repositories.LevelRepository, config repositories.ConfigRepository, party repositories.PartyRepository, boss repositories.BossRepository, log loggo.Logger) Adventure {
	return &adventure{
		areas:     areas,
		classes:   classes,
		users:     users,
		equipment: equips,
		levels:    levels,
		config:    config,
		party:     party,
		damage:    NewDamageService(log),
		boss:      boss,
		log:       log,
	}
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

func (a *adventure) GetBosses(id string) (*[]string, error) {
	var availableBosses []string
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user info: %v", err)
		message := "User has not created an account yet."
		availableBosses = append(availableBosses, message)
		return &availableBosses, nil
	}
	bosses, err := a.boss.QueryDocuments(&[]models.QueryArg{
		{
			Path:  "level",
			Op:    "<=",
			Value: user.ClassMap[user.CurrentClass].Level,
		},
	})
	if err != nil {
		a.log.Errorf("error getting bosses: %v", err)
		message := "There was a problem getting the list of available bosses."
		availableBosses = append(availableBosses, message)
		return &availableBosses, nil
	}
	if bosses == nil {
		message := "There are currently no bosses available to fight you."
		availableBosses = append(availableBosses, message)
		return &availableBosses, nil
	}
	availableBosses = append(availableBosses, fmt.Sprintf("Boss Name:	|	Boss Bonus:	|	Level: "))
	for _, boss := range *bosses {
		availableBosses = append(availableBosses, fmt.Sprintf("%s	|	%s	|	%v", boss.Name, boss.BossBonus.Name, boss.Level))
	}
	return &availableBosses, nil
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
		message := fmt.Sprintf("You are currently in a party!  To leave your current party, first run **!latale -leaveParty**.")
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
		message := fmt.Sprintf("You are currently the leader of a party!  To disband your current party, first run **!latale -leaveParty**.")
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
	message := fmt.Sprintf("**Congratulations**, your party has been created!\nPlease have members join your party by private messaging the command \"***!latale -joinParty %s***\" to NiceHat.\nTo keep players you do not want to join the party from using the command, please do not post the command publicly", *partyId)
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
	if len(party.Members) == 4 {
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
	message := fmt.Sprintf("You have successfully been added to the party!  To leave the party in the future, simply run the command **!latale -leaveParty**")
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
		message := fmt.Sprintf("The %s class does not exist.  Please select a valid class with a valid weapon", user.CurrentClass)
		return &message, nil
	}

	if user.ClassMap[class] == nil && weapon != nil {

		if classInfo.Tier == 1 {
			for _, classWeapon := range classInfo.Weapons {
				if classWeapon.Name == strings.Title(strings.ToLower(*weapon)) {
					user.ClassMap[class] = &models.ClassInfo{
						Name:          classInfo.Name,
						Level:         classInfo.LevelRequirement,
						Exp:           0,
						CurrentWeapon: strings.Title(strings.ToLower(*weapon)),
						Equipment: models.Equipment{
							Weapon: 0,
							Body:   0,
							Glove:  0,
							Shoes:  0,
						},
					}
					user.CurrentClass = class
					_, err := a.users.UpdateDocument(user.ID, user)
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
			for _, learnedClass := range user.ClassMap {
				if learnedClass.Name == *classInfo.ClassRequirement && learnedClass.Level >= classInfo.LevelRequirement {
					message, err := a.ClassAdvance(id, strings.Title(strings.ToLower(*weapon)), class, &learnedClass.Name)
					if err != nil {
						a.log.Errorf("error while switching and upgrading a class: %v", err)
						return nil, err
					}
					return message, nil
				}
			}
		}
	}
	message := "Please provide a weapon, as this is your first time creating this class."
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
	equipmentMap, err := a.getEquipmentMap(user.ClassMap[user.CurrentClass].Equipment)
	if err != nil {
		a.log.Errorf("error getting equipment map: %v", err)
		return nil, nil, err
	}
	a.log.Debugf("equipmentMap: %v", equipmentMap)
	//3.  Use calculateBaseStat method to get stats
	currentStats := a.calculateBaseStat(*user, class.Stats, equipmentMap)
	return &currentStats, nil, nil
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
	for _, wep := range classInfo.Weapons {
		if wep.Name == weapon {
			if *classInfo.ClassRequirement == user.CurrentClass || *classInfo.ClassRequirement == *givenClass {
				classToUse := user.CurrentClass
				if givenClass != nil {
					classToUse = *givenClass
				}
				if classInfo.LevelRequirement <= user.ClassMap[user.CurrentClass].Level {
					user.ClassMap[class] = &models.ClassInfo{
						Name:          classInfo.Name,
						Level:         user.ClassMap[classToUse].Level,
						Exp:           user.ClassMap[classToUse].Exp,
						CurrentWeapon: weapon,
						Equipment:     a.determineStartingGear(classInfo.Tier, user.ClassMap[classToUse].Equipment),
						BossBonuses:   user.ClassMap[classToUse].BossBonuses,
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
		message := fmt.Sprintf("Upon reaching a Second Tier Class, you have obtained the ability to equip the following items: **Bindi, Glasses, Earrings, Ring, Cloak, and Stockings**.\n")
		message += fmt.Sprintf("Your weapon has also been upgraded, and more upgrades have become accessible as a result.  Your continued patronage is appreciated.\n")
		return message
	}
	return ""
}

func (a *adventure) determineStartingGear(tier int32, currentEquips models.Equipment) models.Equipment {
	weapon := 4
	if tier == 3 {
		weapon = 9
	} else if tier == 4 {
		weapon = 12
	}
	bindi := 4
	glasses := 4
	earring := 4
	ring := 4
	mantle := 4
	stocking := 4

	if currentEquips.Bindi != nil {
		bindi = *currentEquips.Bindi
	}
	if currentEquips.Glasses != nil {
		glasses = *currentEquips.Glasses
	}
	if currentEquips.Earring != nil {
		earring = *currentEquips.Earring
	}
	if currentEquips.Ring != nil {
		ring = *currentEquips.Ring
	}
	if currentEquips.Cloak != nil {
		mantle = *currentEquips.Cloak
	}
	if currentEquips.Stockings != nil {
		stocking = *currentEquips.Stockings
	}

	return models.Equipment{
		Weapon:    weapon,
		Body:      currentEquips.Body,
		Glove:     currentEquips.Glove,
		Shoes:     currentEquips.Shoes,
		Bindi:     &bindi,
		Glasses:   &glasses,
		Earring:   &earring,
		Ring:      &ring,
		Cloak:     &mantle,
		Stockings: &stocking,
	}
}

func (a *adventure) UpdateEquipmentPiece(id, equipment string) (*string, error) {
	if equipment == "gloves" {
		equipment = "glove"
	}
	if equipment == "earring" {
		equipment = "earrings"
	}
	if equipment == "cloak" {
		equipment = "mantle"
	}
	//1.  Get User Info
	user, err := a.users.ReadDocument(id)
	if err != nil {
		return nil, err
	}

	//2.  Based on equipment piece, pass current gear level to ProcessUpgrade
	message, err := a.processUpgrade(user, strings.ToLower(equipment))
	if err != nil {
		a.log.Errorf("error processing upgrade: %v", err)
		return nil, err
	}
	return message, nil
}

func (a *adventure) GetEquipmentPieceCost(id, equipment string) (*string, error) {
	if equipment == "gloves" {
		equipment = "glove"
	}
	if equipment == "earring" {
		equipment = "earrings"
	}
	if equipment == "cloak" {
		equipment = "mantle"
	}
	user, err := a.users.ReadDocument(id)
	if err != nil {
		return nil, err
	}
	var equipmentInterface map[string]interface{}
	equips := user.ClassMap[user.CurrentClass].Equipment
	bytes, _ := json.Marshal(&equips)
	json.Unmarshal(bytes, &equipmentInterface)
	equip := equipmentInterface[equipment]
	a.log.Debugf("equips :%v", equip)
	if equip == nil {
		message := fmt.Sprintf("%s is not a valid piece of equipment or you have not unlocked this slot yet!", equipment)
		return &message, nil
	}
	oldEquipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
		return &message, nil
	}
	equipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)+1.0))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
		return &message, nil
	}
	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		message := "Something happened while trying to get class info..."
		return &message, nil
	}
	if classInfo.Tier < equipSheet.TierRequirement {
		message := "You are not at the proper class advancement to advance this piece of equipment any further!"
		return &message, nil
	}

	message := ""
	switch equipment {
	case "bindi":
		oldBindi := strconv.FormatFloat(oldEquipSheet.BindiHP, 'f', -1, 64)
		bindi := strconv.FormatFloat(equipSheet.BindiHP, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("HP gained from bindi: **%s** -> **%s**\n", oldBindi, bindi)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "glasses":
		oldGlasses := strconv.FormatFloat(oldEquipSheet.GlassesCritDamage*100.0, 'f', -1, 64)
		glasses := strconv.FormatFloat(equipSheet.GlassesCritDamage*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("Critical Damage gained from glasses: **%s%%** -> **%s%%**\n", oldGlasses, glasses)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "earrings", "earring":
		oldEarrings := strconv.FormatFloat(oldEquipSheet.EarringCritRate*100.0, 'f', -1, 64)
		earrings := strconv.FormatFloat(equipSheet.EarringCritRate*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("Critical Rate gained from earrings: **%s%%** -> **%s%%**\n", oldEarrings, earrings)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "ring":
		oldRing := strconv.FormatFloat(oldEquipSheet.RingCritRate*100.0, 'f', -1, 64)
		ring := strconv.FormatFloat(equipSheet.RingCritRate*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("Critical Rate gained from ring: **%s%%** -> **%s%%**\n", oldRing, ring)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "cloak", "mantle":
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("Damage gained from cloak: **%1.f** -> **%1.f**\n", oldEquipSheet.MantleDamage, equipSheet.MantleDamage)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "stockings":
		oldStockingsEvasion := strconv.FormatFloat(oldEquipSheet.StockingEvasion*100.0, 'f', -1, 64)
		stockingEvasion := strconv.FormatFloat(equipSheet.StockingEvasion*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.AccessoryCost))
		message += fmt.Sprintf("Evasion gained from stockings: **%s%%** -> **%s%%**\n", oldStockingsEvasion, stockingEvasion)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "weapon":
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
		message += fmt.Sprintf("Damage gained from weapon: **%1.f** -> **%1.f**\n", oldEquipSheet.WeaponDPS, equipSheet.WeaponDPS)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "shoe", "shoes", "boot", "boots":
		oldShoeEvasion := strconv.FormatFloat(oldEquipSheet.ShoeEvasion*100.0, 'f', -1, 64)
		shoeEvasion := strconv.FormatFloat(equipSheet.ShoeEvasion*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
		message += fmt.Sprintf("Evasion gained from shoes: **%s%%** -> **%s%%**\n", oldShoeEvasion, shoeEvasion)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "glove", "gloves":
		oldGloveAccuracy := strconv.FormatFloat(oldEquipSheet.GloveAccuracy*100.0, 'f', -1, 64)
		gloveAccuracy := strconv.FormatFloat(equipSheet.GloveAccuracy*100.0, 'f', -1, 64)
		oldGloveCritDamage := strconv.FormatFloat(oldEquipSheet.GloveCriticalDamage*100.0, 'f', -1, 64)
		gloveCritDamage := strconv.FormatFloat(equipSheet.GloveCriticalDamage*100.0, 'f', -1, 64)
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
		message += fmt.Sprintf("Accuracy gained from gloves: **%s%%** -> **%s%%**\n", oldGloveAccuracy, gloveAccuracy)
		message += fmt.Sprintf("Critical Damage gained from gloves: **%s%%** -> **%s%%**\n", oldGloveCritDamage, gloveCritDamage)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	case "body":
		message += fmt.Sprintf("The cost of upgrading your %s is %s ely.\n", equipment, utils.String(equipSheet.Cost))
		message += fmt.Sprintf("Defense gained from body armor: **%1.f** -> **%1.f**\n", oldEquipSheet.ArmorDefense, equipSheet.ArmorDefense)
		message += fmt.Sprintf("Level requirement: %v", equipSheet.LevelRequirement)
	}
	return &message, nil
}

func (a *adventure) processUpgrade(user *models.User, equipment string) (*string, error) {
	var equipmentInterface map[string]interface{}
	equips := user.ClassMap[user.CurrentClass].Equipment
	a.log.Debugf("equipment: %v", equipment)
	bytes, _ := json.Marshal(&equips)
	json.Unmarshal(bytes, &equipmentInterface)
	equip := equipmentInterface[equipment]
	a.log.Debugf("equips :%v", equip)
	if equip == nil {
		message := fmt.Sprintf("%s is not a valid piece of equipment, or you have not unlocked this slot yet!", equipment)
		return &message, nil
	}

	//3.  Check if an upgrade is available.  If no doc found, gear does not exist, or something happened.
	equipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)+1.0))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
		message := fmt.Sprintf("No further %s upgrades available at this time!", equipment)
		return &message, nil
	}

	if equipSheet.LevelRequirement > user.ClassMap[user.CurrentClass].Level {
		message := fmt.Sprintf("You do not meet the level requirement to upgrade this piece of gear!  Required Level: %v", equipSheet.LevelRequirement)
		return &message, nil
	}
	classInfo, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		message := "Something happened while trying to get class info..."
		return &message, nil
	}
	if classInfo.Tier < equipSheet.TierRequirement {
		message := "You are not at the proper class advancement to advance this piece of equipment any further!"
		return &message, nil
	}

	currentEquipSheet, err := a.equipment.ReadDocument(fmt.Sprintf("%1.f", equip.(float64)))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with error: %v", err)
		return nil, err
	}
	//If the cost is met, decrease user ely by cost, and upgrade specified piece of equipment.

	s := ""
	cost := int32(0)
	switch equipment {
	case "glove", "gloves", "body", "shoe", "shoes", "weapon":
		cost = equipSheet.Cost
		break
	default:
		cost = equipSheet.AccessoryCost
	}
	if *user.Ely >= cost {
		currentValue := equipmentInterface[equipment].(float64)
		currentValue++
		equipmentInterface[equipment] = currentValue
		bytes, _ = json.Marshal(&equipmentInterface)
		var newEquips models.Equipment
		json.Unmarshal(bytes, &newEquips)
		a.log.Debugf("newEquips :%v", newEquips)
		userClass := user.ClassMap[user.CurrentClass]
		userClass.Equipment = newEquips
		ely := *user.Ely
		ely -= cost
		user.Ely = &ely
		user.ClassMap[user.CurrentClass] = userClass
		_, err := a.users.UpdateDocument(user.ID, user)
		if err != nil {
			a.log.Errorf("error updating user document: %v", err)
			return nil, err
		}
		if equipment == "weapon" {
			s = fmt.Sprintf("Successfully upgraded %s from %s to %s!", equipment, currentEquipSheet.WeaponMap[user.ClassMap[user.CurrentClass].CurrentWeapon], equipSheet.WeaponMap[user.ClassMap[user.CurrentClass].CurrentWeapon])
		} else {
			s = fmt.Sprintf("Successfully upgraded %s from %s to %s!", equipment, currentEquipSheet.Name, equipSheet.Name)
		}
	} else {
		s = fmt.Sprintf("Insufficient Ely!  You need %v more Ely to complete this upgrade.", cost-*user.Ely)
		return &s, nil
	}
	return &s, nil
}

func (a *adventure) GetUserInfo(id string) (*models.User, *string, error) {
	a.log.Debugf("id: %s", id)
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		s := "user with this id not found"
		return nil, &s, nil
	}
	a.log.Errorf("userClassMap: %v", user.ClassMap)
	a.log.Errorf("userClassMapEquipment %v", user.ClassMap[user.CurrentClass].Equipment)
	for _, class := range user.ClassMap {
		classEquips := user.ClassMap[class.Name].Equipment
		a.log.Errorf("classEquips %v", classEquips)
		classEquipmentMap, err := a.getEquipmentMap(classEquips)
		if err != nil {
			a.log.Errorf("error getting equipment map: %v", err)
			return nil, nil, err
		}
		var classEquipmentList []string
		classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Body)].Name+" Hat, Shirt, and Pants")
		classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Glove)].Name+" Gloves")
		classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Shoes)].Name+" Shoes")
		classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Weapon)].WeaponMap[user.ClassMap[class.Name].CurrentWeapon])
		jobClass, err := a.classes.ReadDocument(class.Name)
		if err != nil {
			a.log.Errorf("error retrieving class info :%v", err)
			return nil, nil, err
		}
		if jobClass.Tier > 1 {
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Bindi)].Name+" Bindi")
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Glasses)].Name+" Glasses")
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Earring)].Name+" Earrings")
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Ring)].Name+" Ring")
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Cloak)].Name+" Cloak")
			classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(*classEquips.Stockings)].Name+" Stockings")
		} else {
			for i := 0; i < 7; i++ {
				classEquipmentList = append(classEquipmentList, "N/A")
			}
		}

		a.log.Errorf("classEquipmentList: %v", classEquipmentList)
		classInfo := user.ClassMap[class.Name]
		classInfo.Equipment.EquipmentNames = classEquipmentList
		user.ClassMap[class.Name] = classInfo
	}
	if user.Party != nil {
		party, err := a.party.ReadDocument(*user.Party)
		if err != nil {
			a.log.Errorf("error fetching party info")
			return user, nil, nil
		}
		var partyMemberInfo []string
		partyMemberInfo = append(partyMemberInfo, fmt.Sprintf("\n**Party Info:**"))
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

func (a *adventure) getEquipmentMap(classEquips models.Equipment) (map[string]*models.EquipmentSheet, error) {
	classEquipmentMap := make(map[string]*models.EquipmentSheet)
	//Determine Body
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
		message := "Could not find an area with that code.  Please be sure to use the codes specified in **-areas**."
		return nil, &message, nil
	}

	levelCap, err := a.levels.ReadDocument("levelCap")
	if err != nil {
		a.log.Errorf("error getting current level cap: %v", err)
		return nil, nil, err
	}
	a.log.Debugf("levelCap: %v", levelCap)
	if levelCap.Value <= area.LevelRange.Min {
		levelRestriction := fmt.Sprintf("Area %v is currently inaccessible due to level cap restrictions!", areaId)
		return &[]string{levelRestriction}, nil, nil
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
		if monsterMap[utils.String(monster.Rank)] == nil {
			monsterMap[utils.String(monster.Rank)] = &[]models.Monster{}
		}
		updatedList := *monsterMap[utils.String(monster.Rank)]
		updatedList = append(updatedList, monster)
		monsterMap[utils.String(monster.Rank)] = &updatedList
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
	adventureLog, err := a.createAdventureLog(*classInfo, user, *currentStats, monster)
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
	adventureLog, err = a.partyBattleLog(partyMemberInfos, encounteredMonsters, adventureLog, user)
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
		if monsterMap[utils.String(monster.Rank)] == nil {
			monsterMap[utils.String(monster.Rank)] = &[]models.Monster{}
		}
		updatedList := *monsterMap[utils.String(monster.Rank)]
		updatedList = append(updatedList, monster)
		monsterMap[utils.String(monster.Rank)] = &updatedList
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
			Name:         monster.Name + " " + string('A'+i) + monsterRank,
			Ely:          monster.Ely,
			Exp:          monster.Exp,
		})
	}
	return encounteredMonsters, nil
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
		partyMemberInfos = append(partyMemberInfos, &models.UserBlob{
			User:         userInfo,
			StatModifier: userStats,
			JobClass:     userJob,
			CurrentHP:    int(userStats.HP),
			MaxHP:        int(userStats.HP),
			UserLevel:    userInfo.ClassMap[userInfo.CurrentClass].Level,
			Weapon:       userInfo.ClassMap[userInfo.CurrentClass].CurrentWeapon,
		})
	}
	return partyMemberInfos, nil
}

func (a *adventure) checkAdventureCooldown(user *models.User, boss bool) (*string, bool) {
	var lastAction time.Time
	var commandType string
	if boss {
		lastAction = user.LastBossActionTime.Add(45 * time.Minute)
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
		message := "You must be in a party to participate in Boss Fights!  Join a party, or create one using `!latale -createParty`."
		return nil, &message, nil
	}
	boss, err := a.boss.ReadDocument(bossId)
	if err != nil {
		a.log.Errorf("error getting area info: %v", err)
		message := "Could not find a boss with that name.  Please be sure to use the names specified in **-bosses**."
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
	}
	if len(coolDownLog) > 0 {
		adventureLog = append(adventureLog, coolDownLog)
	}
	if len(adventureLog) != 0 {
		return &adventureLog, nil, nil
	}
	adventureLog, err = a.bossBattleLog(partyMembers, boss, user.ID)
	if err != nil {
		a.log.Errorf("error while generating boss log: %v", err)
		return nil, nil, err
	}
	return &adventureLog, nil, nil
}

func (a *adventure) bossBattleLog(users []*models.UserBlob, boss *models.Monster, primaryUserId string) ([]string, error) {
	battleWin := false
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	bossMaxHp := int(boss.Stats.HP) * len(users)
	bossCurrentHp := bossMaxHp
	partyBonus := 1.0
	if len(users) > 1 {
		partyBonus = (float64(len(users)) / 10.0) + 1.0
	}
	bossHPPercentage := float64(bossCurrentHp) / float64(bossMaxHp)
	var adventureLog []string
	//Set cooldowns of all boss skills to 0 for beginning.
	for _, skill := range *boss.Skills {
		curCd := int32(0)
		skill.CurrentCoolDown = &curCd
	}
	activeSkills := 1
	enraged := false
	adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BOSS ENCOUNTER------------------------**\n__**The Party**__ has encountered **%s**, __**%s**__.\n**------------------------BOSS ENCOUNTER------------------------**", boss.Name, *boss.BossTitle))
bossBattle:
	for len(users) != 0 && bossCurrentHp != 0 {
		userLogs := ""
		for _, user := range users {
			if user.CrowdControlled != nil && *user.CrowdControlled != 0 && *user.CrowdControlStatus != "poisoned" {
				userLogs += fmt.Sprintf("__**%s**__ is currently %s for **%v turn(s)**.\n", user.User.Name, *user.CrowdControlStatus, *user.CrowdControlled)
			} else {
				userLog, damage := a.damage.DetermineHit(randGenerator, user.User.Name, boss.Name, *user.StatModifier, boss.Stats, &user.Weapon, user.JobClass, &user.UserLevel, true)
				bossCurrentHp = ((int(bossCurrentHp) - int(damage)) + int(math.Abs(float64(bossCurrentHp-damage)))) / 2
				userLogs += userLog + "\n"
				bossHPPercentage = float64(bossCurrentHp) / float64(bossMaxHp)
				userLogs += fmt.Sprintf("__**%s**__'s **HP: %s%%/100%%**\n", boss.Name, fmt.Sprintf("%.2f", bossHPPercentage*100))
			}
			if bossCurrentHp <= 0 {
				userLogs += fmt.Sprintf("__**The Party**__ **has successfully defeated ** __**%s**__!\n", boss.Name)
				adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN PARTY ATTACK TURN------------------------**\n%s**------------------------END PARTY ATTACK TURN------------------------**", userLogs))
				battleWin = true
				break bossBattle
			}
		}
		adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN PARTY ATTACK TURN------------------------**\n%s**------------------------END PARTY ATTACK TURN------------------------**", userLogs))
		phase, newActiveSkills := a.checkPhaseStatus(bossHPPercentage, boss, activeSkills)
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
		bossLogs := ""
		if active > *boss.IdleTime || enraged {
			skills := *boss.Skills
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
					updatedUser, bossDamageLog, damage := a.damage.DetermineBossDamage(randGenerator, *user, boss, skill)
					updatedUser.CurrentHP = ((updatedUser.CurrentHP - int(damage)) + int(math.Abs(float64(updatedUser.CurrentHP-damage)))) / 2
					bossLogs += bossDamageLog + "\n"
					bossLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", updatedUser.User.Name, updatedUser.CurrentHP, updatedUser.MaxHP)
					alivePlayers[i] = updatedUser
					if updatedUser.CurrentHP <= 0 {
						bossLogs += fmt.Sprintf("**%s was killed by %s!**\n", user.User.Name, boss.Name)
					}
				}
				users = []*models.UserBlob{}
				for _, user := range alivePlayers {
					if user.CurrentHP > 0 {
						users = append(users, user)
					}
				}
			} else {
				targetedUser := randGenerator.Intn(len(users))
				updatedUser, bossDamageLog, damage := a.damage.DetermineBossDamage(randGenerator, *users[targetedUser], boss, skill)
				updatedUser.CurrentHP = ((updatedUser.CurrentHP - int(damage)) + int(math.Abs(float64(updatedUser.CurrentHP-damage)))) / 2
				bossLogs += bossDamageLog + "\n"
				bossLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", updatedUser.User.Name, updatedUser.CurrentHP, updatedUser.MaxHP)
				if updatedUser.CurrentHP <= 0 {
					bossLogs += fmt.Sprintf("**%s was killed by %s!**\n", updatedUser.User.Name, boss.Name)
					copy(users[targetedUser:], users[targetedUser+1:]) // Shift a[i+1:] left one index.
					users[len(users)-1] = nil                          // Erase last element (write zero value).
					users = users[:len(users)-1]
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
			boss.Skills = &skills
		} else {
			bossLogs += *boss.IdlePhrase + "\n"
		}
		var alivePlayers []*models.UserBlob
		for _, user := range users {
			alivePlayers = append(alivePlayers, user)
		}
		playerDied := false
		for i, user := range users {
			if user.CurrentHP > 0 && user.CrowdControlled != nil && *user.CrowdControlled != 0 && *user.CrowdControlStatus == "poisoned" {
				poisonDamage := int(float64(user.MaxHP) * 0.05)
				user.CurrentHP = ((user.CurrentHP - int(poisonDamage)) + int(math.Abs(float64(user.CurrentHP-poisonDamage)))) / 2
				bossLogs += fmt.Sprintf("**%s lost %v HP!** due to being **%s**. %s is **%s** for **%v turn(s)**.\n", user.User.Name, poisonDamage, *user.CrowdControlStatus, user.User.Name, *user.CrowdControlStatus, *user.CrowdControlled)
				bossLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", user.User.Name, user.CurrentHP, user.MaxHP)
				alivePlayers[i] = user
				if user.CurrentHP <= 0 {
					bossLogs += fmt.Sprintf("**%s was killed by %s!**\n", user.User.Name, boss.Name)
					playerDied = true
				}
			}
		}
		if playerDied {
			for _, user := range alivePlayers {
				if user.CurrentHP > 0 {
					users = append(users, user)
				}
			}
		}
		adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN BOSS ATTACK TURN------------------------**\n%s**------------------------END BOSS ATTACK TURN------------------------**", bossLogs))

		healLogs := ""
		for i, user := range users {
			if user.CrowdControlled != nil && *user.CrowdControlled > 0 {
				crowdControlled := *user.CrowdControlled
				crowdControlled--
				user.CrowdControlled = &crowdControlled
				if *user.CrowdControlled == 0 {
					user.CrowdControlStatus = nil
				}
			}
			if user.CurrentHP != int(user.StatModifier.HP) && user.StatModifier.Recovery > 0.0 {
				userHeal := int(user.StatModifier.HP * user.StatModifier.Recovery)
				if userHeal+user.CurrentHP > int(user.StatModifier.HP) {
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
			userInfo := winningUsers.User
			if levelCap.Value > winningUsers.User.ClassMap[winningUsers.User.CurrentClass].Level {
				userClassInfo := *winningUsers.User.ClassMap[winningUsers.User.CurrentClass]
				userClassInfo, adventureLog = a.determineBossBonusDrop(winningUsers.User.Name, userClassInfo, *boss, adventureLog)
				bossExp := int32(float64(boss.Exp*int32(*expGainRate)) * partyBonus)
				userClassInfo.Exp += bossExp
				oldEly := *winningUsers.User.Ely
				bossEly := int32(float64(boss.Ely*int32(*expGainRate)) * partyBonus)
				oldEly += bossEly
				userInfo.ClassMap[userInfo.CurrentClass] = &userClassInfo
				userInfo.Ely = &oldEly
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", winningUsers.User.Name, utils.String(bossExp), bossEly))
				newUserClassInfo, newAdventureLog, err := a.processLevelUps(userClassInfo, adventureLog, winningUsers.User, levelCap.Value)
				if err != nil {
					a.log.Errorf("error processing level ups: %v", err)
					return adventureLog, nil
				}
				winningUsers.User.ClassMap[winningUsers.User.CurrentClass] = &newUserClassInfo
				adventureLog = newAdventureLog
			} else if battleWin && levelCap.Value == winningUsers.User.ClassMap[winningUsers.User.CurrentClass].Level {
				userClassInfo := *winningUsers.User.ClassMap[winningUsers.User.CurrentClass]
				userClassInfo, adventureLog = a.determineBossBonusDrop(winningUsers.User.Name, userClassInfo, *boss, adventureLog)
				bossExp := int32(float64(boss.Exp*int32(*expGainRate)) * partyBonus)
				userClassInfo.Exp += bossExp
				oldEly := *winningUsers.User.Ely
				bossEly := int32(float64(boss.Ely*int32(*expGainRate)) * partyBonus)
				oldEly += bossEly
				userInfo.ClassMap[userInfo.CurrentClass] = &userClassInfo
				userInfo.Ely = &oldEly
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", winningUsers.User.Name, utils.String(bossExp), bossEly))
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ has hit the current Level Cap of: %v, and can no longer level up.", winningUsers.User.Name, levelCap.Value))
			}
			userInfo.LastBossActionTime = time.Now()

			_, err := a.users.UpdateDocument(userInfo.ID, userInfo)
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

func (a *adventure) determineBossBonusDrop(userName string, userClassInfo models.ClassInfo, boss models.Monster, adventureLog []string) (models.ClassInfo, []string) {
	dropChance := rand.Float64()
	if userClassInfo.BossBonuses[boss.Name] == nil && dropChance <= *boss.BossBonus.BossDropChance {
		if userClassInfo.BossBonuses == nil {
			userClassInfo.BossBonuses = make(map[string]*models.BossBonus)
		}
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained the ability __**%s**__ for defeating __**%s**__ on their current class!", userName, boss.BossBonus.Name, boss.Name))
		userClassInfo.BossBonuses[boss.Name] = boss.BossBonus
	}
	return userClassInfo, adventureLog
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

func (a *adventure) partyBattleLog(users []*models.UserBlob, encounteredMonsters []*models.MonsterBlob, adventureLog []string, primaryUser *models.User) ([]string, error) {
	battleWin := false
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	totalExpReward := int32(0)
	totalElyReward := int32(0)
	partyBonus := 1.0 + (float64(len(users)/10.0) * 3.5)
	monsterNames := ""
	for _, monster := range encounteredMonsters {
		totalExpReward += int32(float64(monster.Exp) * partyBonus)
		totalElyReward += int32(float64(monster.Ely) * partyBonus)
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
combat:
	for a.checkGroupDeaths(users, encounteredMonsters) {
		//Party will target enemies in order of how they spawn.
		//Enemies will attack party members randomly.
		//Battle continues until one side is no longer able to fight.
		userLogs := ""
		for _, user := range users {
			userLog, damage := a.damage.DetermineHit(randGenerator, user.User.Name, encounteredMonsters[0].Name, *user.StatModifier, *encounteredMonsters[0].StatModifier, &user.Weapon, user.JobClass, &user.UserLevel, false)
			currentMonsterHP := int(encounteredMonsters[0].CurrentHP)
			currentMonsterHP = ((int(currentMonsterHP) - int(damage)) + int(math.Abs(float64(currentMonsterHP-damage)))) / 2
			userLogs += userLog + "\n"
			monsterMaxHp := int(encounteredMonsters[0].StatModifier.HP)
			userLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v\n", encounteredMonsters[0].Name, currentMonsterHP, monsterMaxHp)
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
		adventureLog = append(adventureLog, fmt.Sprintf("%s", userLogs))
		enemiesLog := ""
		for _, monster := range encounteredMonsters {
			a.log.Debugf("users: %v", users)
			targetedUser := randGenerator.Intn(len(users))
			monsterLog, damage := a.damage.DetermineHit(randGenerator, monster.Name, users[targetedUser].User.Name, *monster.StatModifier, *users[targetedUser].StatModifier, nil, nil, nil, false)
			users[targetedUser].CurrentHP = ((users[targetedUser].CurrentHP - int(damage)) + int(math.Abs(float64(users[targetedUser].CurrentHP-damage)))) / 2
			enemiesLog += "				" + monsterLog + "\n"
			enemiesLog += fmt.Sprintf("				__**%s**__'s HP: %v/%v\n", users[targetedUser].User.Name, users[targetedUser].CurrentHP, users[targetedUser].MaxHP)
			if users[targetedUser].CurrentHP <= 0 {
				enemiesLog += fmt.Sprintf("				**%s was killed by %s!**\n", users[targetedUser].User.Name, monster.Name)
				copy(users[targetedUser:], users[targetedUser+1:]) // Shift a[i+1:] left one index.
				users[len(users)-1] = nil                          // Erase last element (write zero value).
				users = users[:len(users)-1]
			}
			a.log.Debugf("users: %v", users)
			if len(users) == 0 {
				adventureLog = append(adventureLog, fmt.Sprintf("%s", enemiesLog))
				break combat
			}
		}
		adventureLog = append(adventureLog, fmt.Sprintf("%s", enemiesLog))
		healLogs := ""
		for i, user := range users {
			userHeal := int(user.StatModifier.HP * user.StatModifier.Recovery)
			if user.CurrentHP != int(user.StatModifier.HP) && user.StatModifier.Recovery > 0.0 {
				if userHeal+user.CurrentHP > int(user.StatModifier.HP) {
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
				enemyHealLogs += fmt.Sprintf("__**%s**__ **HEALED** for %v HP.\n", monster.Name, monsterHeal)
				enemyHealLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", monster.Name, monsterHeal+monster.CurrentHP, strconv.FormatFloat(monster.StatModifier.HP, 'f', -1, 64))
			}
		}
		if enemyHealLogs != "" {
			adventureLog = append(adventureLog, fmt.Sprintf("**------------------------BEGIN ENEMIES HEAL TURN------------------------**\n%s**------------------------END ENEMIES HEAL TURN------------------------**", enemyHealLogs))

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
			userInfo := user.User
			if levelCap.Value > user.User.ClassMap[user.User.CurrentClass].Level {
				userClassInfo := *user.User.ClassMap[user.User.CurrentClass]
				monsterExp := totalExpReward / int32(len(users)) * int32(*expGainRate)
				userClassInfo.Exp += monsterExp
				monsterEly := totalElyReward / int32(len(users)) * int32(*expGainRate)
				oldEly := *user.User.Ely
				oldEly += totalElyReward / int32(len(users)) * int32(*expGainRate)
				userInfo.Ely = &oldEly
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", user.User.Name, utils.String(monsterExp), monsterEly))
				newUserClassInfo, newAdventureLog, err := a.processLevelUps(userClassInfo, adventureLog, user.User, levelCap.Value)
				if err != nil {
					a.log.Errorf("error processing level ups: %v", err)
					return adventureLog, nil
				}
				user.User.ClassMap[user.User.CurrentClass] = &newUserClassInfo
				adventureLog = newAdventureLog
			} else if battleWin && levelCap.Value == user.User.ClassMap[user.User.CurrentClass].Level {
				userClassInfo := *user.User.ClassMap[user.User.CurrentClass]
				monsterExp := totalExpReward / int32(len(users)) * int32(*expGainRate)
				userClassInfo.Exp += monsterExp
				monsterEly := totalElyReward / int32(len(users)) * int32(*expGainRate)
				oldEly := *user.User.Ely
				oldEly += totalElyReward / int32(len(users)) * int32(*expGainRate)
				userInfo.ClassMap[userInfo.CurrentClass] = &userClassInfo
				userInfo.Ely = &oldEly
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%v*** Ely!", user.User.Name, utils.String(monsterExp), monsterEly))
				adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ has hit the current Level Cap of: %v, and can no longer level up.", user.User.Name, levelCap.Value))
			}
			if primaryUser.ID == userInfo.ID {
				userInfo.LastActionTime = time.Now()
			}
			_, err := a.users.UpdateDocument(userInfo.ID, userInfo)
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

func (a *adventure) createAdventureLog(classInfo models.JobClass, user *models.User, userStats models.StatModifier, monster models.Monster) ([]string, error) {
	var adventureLog []string
	logLine, onCooldown := a.checkAdventureCooldown(user, false)
	if onCooldown {
		adventureLog = append(adventureLog, *logLine)
		return adventureLog, nil
	}
	battleWin := false
	userMaxHP := int(userStats.HP)
	monsterMaxHp := int(monster.Stats.HP)
	currentHP := int(userStats.HP)
	monsterHP := int(monster.Stats.HP)
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	rankExclamation := ""
	for i := int32(0); i < monster.Rank; i++ {
		rankExclamation += "!"
	}
	adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ has encountered a __**%s**__**%s**", user.Name, monster.Name, rankExclamation))
	userLevel := user.ClassMap[user.CurrentClass].Level
	userWeapon := user.ClassMap[user.CurrentClass].CurrentWeapon
	for currentHP != 0 && monsterHP != 0 {
		userLog, damage := a.damage.DetermineHit(randGenerator, user.Name, monster.Name, userStats, monster.Stats, &userWeapon, &classInfo, &userLevel, false)
		monsterHP = ((int(monsterHP) - int(damage)) + int(math.Abs(float64(monsterHP-damage)))) / 2
		userLog += fmt.Sprintf("\n__**%s**__'s HP: %v/%v\n", monster.Name, monsterHP, monsterMaxHp)
		if monsterHP <= 0 {
			userLog += fmt.Sprintf("__**%s**__ **has successfully defeated the** __**%s**__!\n", user.Name, monster.Name)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", userLog))
			battleWin = true
			break
		}
		adventureLog = append(adventureLog, fmt.Sprintf("%s", userLog))

		monsterLog, damage := a.damage.DetermineHit(randGenerator, monster.Name, user.Name, monster.Stats, userStats, nil, nil, nil, false)
		currentHP = ((int(currentHP) - int(damage)) + int(math.Abs(float64(currentHP-damage)))) / 2
		monsterLog = "				" + monsterLog
		monsterLog += fmt.Sprintf("\n				__**%s**__'s HP: %v/%v\n", user.Name, currentHP, userMaxHP)
		if currentHP <= 0 {
			monsterLog += fmt.Sprintf("				**%s was killed by %s!**\n", user.Name, monster.Name)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", monsterLog))
			break
		}
		adventureLog = append(adventureLog, fmt.Sprintf("%s", monsterLog))

		userHeal := int(userStats.HP * userStats.Recovery)
		if userStats.Recovery > 0.0 && currentHP != int(userStats.HP) {
			healLogs := ""
			if userHeal+currentHP > int(userStats.HP) {
				currentHP = int(userStats.HP)
			} else {
				currentHP += userHeal
			}
			healLogs += fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.\n", user.Name, userHeal)
			healLogs += fmt.Sprintf("__**%s**__'s HP: %v/%v!\n", user.Name, currentHP, userMaxHP)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", healLogs))
		}

		if monster.Stats.Recovery > 0.0 {
			healLogs := ""
			monsterHeal := int(monster.Stats.HP * monster.Stats.Recovery)
			if monsterHeal+monsterHP > int(monster.Stats.HP) {
				monsterHP = int(monster.Stats.HP)
			} else {
				monsterHP += monsterHeal
			}
			healLogs += fmt.Sprintf("				__**%s**__ **HEALED** for %v HP.\n", monster.Name, monsterHeal)
			healLogs += fmt.Sprintf("				__**%s**__'s HP: %v/%v!\n", monster.Name, monsterHP, monsterMaxHp)
			adventureLog = append(adventureLog, fmt.Sprintf("%s", healLogs))
		}

	}
	levelCap, err := a.levels.ReadDocument("levelCap")
	if err != nil {
		a.log.Errorf("error retrieving current levelCap: %v", err)
		return nil, err
	}
	if battleWin && levelCap.Value > user.ClassMap[user.CurrentClass].Level {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- %s WON THE BATTLE.  GETTING RESULTS. ----------------------------**", user.Name))
		userClassInfo := *user.ClassMap[user.CurrentClass]
		expGainRate, err := a.GetExpGainRate("exp")
		if err != nil {
			return nil, err
		}
		monsterExp := monster.Exp * int32(*expGainRate)
		userClassInfo.Exp += monsterExp
		monsterEly := monster.Ely * int32(*expGainRate)
		*user.Ely += monsterEly
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%s*** Ely!", user.Name, utils.String(monsterExp), utils.String(monsterEly)))
		a.log.Debugf("userClassInfo: %v\n", userClassInfo)
		newUserClassInfo, newAdventureLog, err := a.processLevelUps(userClassInfo, adventureLog, user, levelCap.Value)
		if err != nil {
			a.log.Errorf("error processing level ups: %v", err)
			return adventureLog, nil
		}

		user.ClassMap[user.CurrentClass] = &newUserClassInfo
		adventureLog = newAdventureLog
	} else if battleWin && levelCap.Value == user.ClassMap[user.CurrentClass].Level {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- %s WON THE BATTLE.  GETTING RESULTS. ----------------------------**", user.Name))
		userClassInfo := *user.ClassMap[user.CurrentClass]
		expGainRate, err := a.GetExpGainRate("exp")
		if err != nil {
			return nil, err
		}
		monsterExp := monster.Exp * int32(*expGainRate)
		userClassInfo.Exp += monsterExp
		monsterEly := monster.Ely * int32(*expGainRate)
		*user.Ely += monsterEly
		user.ClassMap[user.CurrentClass] = &userClassInfo
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ gained ***%s*** points of experience and ***%s*** Ely!", user.Name, utils.String(monsterExp), utils.String(monsterEly)))
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ has hit the current Level Cap of: %v, and can no longer level up.", user.Name, levelCap.Value))
	} else {
		adventureLog = append(adventureLog, fmt.Sprintf("**---------------------------- %s LOST THE BATTLE. ----------------------------**", user.Name))
	}
	user.LastActionTime = time.Now()
	_, err = a.users.UpdateDocument(user.ID, user)
	if err != nil {
		a.log.Errorf("failed to update user doc with error: %v", err)
		return adventureLog, nil
	}
	return adventureLog, nil
}

func (a *adventure) processLevelUps(userClassInfo models.ClassInfo, adventureLog []string, user *models.User, levelCap int32) (models.ClassInfo, []string, error) {
	level, err := a.levels.ReadDocument(utils.String(userClassInfo.Level))
	if err != nil {
		a.log.Errorf("error getting level data: %v", err)
		return userClassInfo, adventureLog, err
	}
	if userClassInfo.Exp >= level.Exp {
		userClassInfo.Exp -= level.Exp
		userClassInfo.Level++
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ **LEVELED UP**!  Current Level: %v", user.Name, userClassInfo.Level))
		if userClassInfo.Level == 50 {

			advanceJobs, err := a.classes.QueryDocuments(&[]models.QueryArg{{Path: "classRequirement", Op: "==", Value: user.CurrentClass}})
			if err != nil {
				a.log.Errorf("error querying for 2nd tier classes")
				return userClassInfo, adventureLog, err
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
			if advanceJobs != nil {
				adventureLog = append(adventureLog, fmt.Sprintf("Congratulations!  Now that you've reached level 50, you may use the **-classAdvance <Class> <Weapon>** command to advance to ***%s***", jobText))
			}
		}
		if levelCap == userClassInfo.Level {
			userClassInfo.Exp = 0.0
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ ** has reached the level cap!**", user.Name))
			return userClassInfo, adventureLog, nil
		}
		return a.processLevelUps(userClassInfo, adventureLog, user, levelCap)
	} else {
		adventureLog = append(adventureLog, fmt.Sprintf("Current Exp: **%s/%s**", utils.String(userClassInfo.Exp), utils.String(level.Exp)))
	}
	return userClassInfo, adventureLog, nil
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

func (a *adventure) addNewEquipmentSheet(equipSheet map[string]*models.EquipmentSheet, equipment *models.EquipmentSheet) map[string]*models.EquipmentSheet {
	if equipSheet[equipment.ID] == nil {
		equipSheet[equipment.ID] = equipment
	}
	return equipSheet
}

func (a *adventure) calculateBaseStat(user models.User, class models.StatModifier, equipmentMap map[string]*models.EquipmentSheet) models.StatModifier {
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
	a.log.Debugf("user bossbonuses: %v", user.ClassMap[user.CurrentClass].BossBonuses)
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
		}
	}
	return models.StatModifier{
		MaxDPS:                 getDynamicStat(20, levelModifier, class.MaxDPS) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].WeaponDPS + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].MantleDamage + bossMaxDps,
		MinDPS:                 getDynamicStat(20, levelModifier, class.MinDPS) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].WeaponDPS + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].MantleDamage + bossMinDps,
		Defense:                getDynamicStat(15, levelModifier, class.Defense) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Body)].ArmorDefense + bossDefense,
		HP:                     getDynamicStat(100, levelModifier, class.HP) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].BindiHP + bossHp,
		Recovery:               getStaticStat(0.05, levelModifier, class.Recovery) + bossRecv,
		CriticalDamageModifier: getStaticStat(1.5, levelModifier, class.CriticalDamageModifier) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Body)].GloveCriticalDamage + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].GlassesCritDamage + bossCritDmg,
		CriticalRate:           getStaticStat(0.05, levelModifier, class.CriticalRate) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].EarringCritRate + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].RingCritRate + bossCritRate,
		SkillProcRate:          getStaticStat(0.25, levelModifier, class.SkillProcRate) + bossSkillProc,
		Evasion:                getStaticStat(0.05, levelModifier, class.Evasion) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Shoes)].ShoeEvasion + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].StockingEvasion + bossEvasion,
		Accuracy:               getStaticStat(0.95, levelModifier, class.Accuracy) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Glove)].GloveAccuracy + bossAccuracy,
		SkillDamageModifier:    class.SkillDamageModifier,
	}
}

func getDynamicStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * statModifier * math.Pow(levelModifier, 7)
}

func getStaticStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * levelModifier * statModifier
}
