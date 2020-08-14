package services

import (
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/repositories"
	"math"
	"strconv"
)

type Adventure interface {
	GetBaseStat(id string) (*models.StatModifier, *string, error)
	GetArea(id string) (*models.Area, error)
	GetUserInfo(id string) (*models.User, *string, error)
}

type adventure struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	users     repositories.UserRepository
	equipment repositories.EquipmentRepository
	log       loggo.Logger
}

func NewAdventureService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, equips repositories.EquipmentRepository, log loggo.Logger) Adventure {
	return &adventure{
		areas:     areas,
		classes:   classes,
		users:     users,
		equipment: equips,
		log:       log,
	}
}

func (a *adventure) GetArea(id string) (*models.Area, error) {
	area, err := a.areas.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting area: %v", err)
		return nil, err
	}
	return area, nil
}

func (a *adventure) GetBaseStat(id string) (*models.StatModifier, *string, error) {
	//1.  Get User Data based on ID
	a.log.Debugf("id: %s", id)
	user, err := a.users.ReadDocument(id)
	if err != nil {
		a.log.Errorf("error getting user stats: %v", err)
		return nil, nil, err
	}
	a.log.Debugf("user: %v", user)
	class, err := a.classes.ReadDocument(user.CurrentClass)
	if err != nil {
		a.log.Errorf("error reading currently selected class")
		return nil, nil, err
	}
	//3.  Use calculateBaseStat method to get stats
	currentStats := a.calculateBaseStat(float64(*user.CurrentLevel), class.Stats)
	return &currentStats, nil, nil
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
	classEquips := user.ClassMap[user.CurrentClass].Equipment
	a.log.Errorf("classEquips %v", classEquips)

	classEquipmentMap := make(map[string]*models.EquipmentSheet)
	var classEquipmentList []string
	//Determine Body
	equipmentSheetBody, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Body))
	if err != nil {
		a.log.Errorf("error retrieving equipment sheet with provided equipment")
		return nil, nil, err
	}
	classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBody)
	classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Body)].Name+" Hat, Shirt, and Pants")
	//Determine Gloves
	if classEquipmentMap[strconv.Itoa(classEquips.Glove)] == nil {
		equipmentSheetBody, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Glove))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBody)
	}
	classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Glove)].Name+" Gloves")
	//Determine Shoes
	if classEquipmentMap[strconv.Itoa(classEquips.Shoes)] == nil {
		equipmentSheetBody, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Shoes))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBody)
	}
	classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Shoes)].Name+" Shoes")
	//Determine WeaponMap
	if classEquipmentMap[strconv.Itoa(classEquips.Weapon)] == nil {
		equipmentSheetBody, err := a.equipment.ReadDocument(strconv.Itoa(classEquips.Weapon))
		if err != nil {
			a.log.Errorf("error retrieving equipment sheet with provided equipment")
			return nil, nil, err
		}
		classEquipmentMap = a.addNewEquipmentSheet(classEquipmentMap, equipmentSheetBody)
	}
	classEquipmentList = append(classEquipmentList, classEquipmentMap[strconv.Itoa(classEquips.Weapon)].WeaponMap[user.CurrentWeapon])

	a.log.Errorf("classEquipmentList: %v", classEquipmentList)
	classInfo := user.ClassMap[user.CurrentClass]
	classInfo.Equipment.EquipmentNames = classEquipmentList
	user.ClassMap[user.CurrentClass] = classInfo
	a.log.Errorf("userclassMap: %v", user.ClassMap[user.CurrentClass])

	return user, nil, nil
}

func (a *adventure) addNewEquipmentSheet(equipSheet map[string]*models.EquipmentSheet, equipment *models.EquipmentSheet) map[string]*models.EquipmentSheet {
	if equipSheet[equipment.ID] == nil {
		equipSheet[equipment.ID] = equipment
	}
	return equipSheet
}

func (a *adventure) calculateBaseStat(level float64, class models.StatModifier) models.StatModifier {
	levelModifier := float64((level / 100) + 1)
	return models.StatModifier{
		DPS:                    getDynamicStat(20, levelModifier, level, class.DPS),
		Defense:                getDynamicStat(15, levelModifier, level, class.Defense),
		HP:                     getDynamicStat(100, levelModifier, level, class.HP),
		Recovery:               getStaticStat(0.05, levelModifier, class.Recovery),
		CriticalDamageModifier: getStaticStat(1.5, levelModifier, class.CriticalDamageModifier),
		CriticalRate:           getStaticStat(0.05, levelModifier, class.CriticalRate),
		SkillProcRate:          getStaticStat(0.25, levelModifier, class.SkillProcRate),
		Evasion:                getStaticStat(0.05, levelModifier, class.Evasion),
		Accuracy:               getStaticStat(0.95, levelModifier, class.Accuracy),
	}
}

func getDynamicStat(baseStat, levelModifier, level, statModifier float64) float64 {
	return baseStat * statModifier * math.Pow(levelModifier, 7)
}

func getStaticStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * levelModifier * statModifier
}
