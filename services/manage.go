package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/repositories"
	"lataleBotService/utils"
	"math"
	"strconv"
	"strings"
)

type Manage interface {
	AddNewUser(user models.User, weapon string) (*string, *string, error)
	AddNewClass(class *models.JobClass) (*string, error)
	AddNewArea(area models.Area) (*string, error)
	AddNewMonster(area *models.Area, monster models.Monster) (*string, error)
	ConvertToInventorySystemBatch() (*string, error)
	AddNewItem(item *models.Item) (*string, error)
	IncreaseLevelCap(level int) (*[]models.Level, error)
	CreateExpTable(levels []models.Level) (*[]models.Level, error)
	ToggleExpEvent(expRate int) error
	AddNewEquipmentSheet(equipment models.OldEquipmentSheet) (*string, error)
	AddNewBoss(boss models.Monster) (*string, error)
}

type manage struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	levels    repositories.LevelRepository
	users     repositories.UserRepository
	equipment repositories.EquipmentRepository
	config    repositories.ConfigRepository
	boss      repositories.BossRepository
	item      repositories.ItemRepository
	log       loggo.Logger
}

func NewManageService(areas repositories.AreasRepository, levels repositories.LevelRepository, classes repositories.ClassRepository, users repositories.UserRepository, equip repositories.EquipmentRepository, config repositories.ConfigRepository, boss repositories.BossRepository, item repositories.ItemRepository, log loggo.Logger) Manage {
	return &manage{
		areas:     areas,
		classes:   classes,
		levels:    levels,
		users:     users,
		equipment: equip,
		config:    config,
		boss:      boss,
		item:      item,
		log:       log,
	}
}

func (m *manage) ConvertToInventorySystemBatch() (*string, error) {
	users, err := m.users.QueryDocuments(nil)
	if err != nil {
		m.log.Errorf("error getting users: %v", err)
		return nil, err
	}
	oldEquipMap := make(map[int]*models.OldEquipmentSheet)
	for i := 0; i < 7; i++ {
		equipmentSheet, err := m.equipment.ReadDocument(strconv.Itoa(i))
		if err != nil {
			m.log.Errorf("error getting equipment sheets")
			return nil, err
		}
		oldEquipMap[i] = equipmentSheet
	}

	m.log.Debugf("users: %v", users)
	for _, user := range *users {
		err := m.updateUserToInventorySystem(&user, oldEquipMap)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (m *manage) updateUserToInventorySystem(user *models.User, oldEquipMap map[int]*models.OldEquipmentSheet) error {
	//Iterate over user classes to begin changing the old equipment to the new type
	copyUser := user
	newClassMap := map[string]*models.ClassInfo{}
	for _, class := range user.ClassMap {
		classInfo := class
		classInfo.Equipment.Weapon = m.changeEquipToNewEquipType("Weapon", class.CurrentWeapon, oldEquipMap[class.OldEquipmentSheet.Weapon])
		classInfo.Equipment.Top = m.changeEquipToNewEquipType("Top", nil, oldEquipMap[class.OldEquipmentSheet.Body])
		classInfo.Equipment.Bottom = m.changeEquipToNewEquipType("Bottom", nil, oldEquipMap[class.OldEquipmentSheet.Body])
		classInfo.Equipment.Headpiece = m.changeEquipToNewEquipType("Headpiece", nil, oldEquipMap[class.OldEquipmentSheet.Body])
		classInfo.Equipment.Glove = m.changeEquipToNewEquipType("Gloves", nil, oldEquipMap[class.OldEquipmentSheet.Glove])
		classInfo.Equipment.Shoes = m.changeEquipToNewEquipType("Boots", nil, oldEquipMap[class.OldEquipmentSheet.Shoes])
		if class.OldEquipmentSheet.Bindi != nil {
			bindi := m.changeEquipToNewEquipType("Bindi", nil, oldEquipMap[*class.OldEquipmentSheet.Bindi])
			classInfo.Equipment.Bindi = &bindi
			glasses := m.changeEquipToNewEquipType("Glasses", nil, oldEquipMap[*class.OldEquipmentSheet.Glasses])
			classInfo.Equipment.Glasses = &glasses
			earrings := m.changeEquipToNewEquipType("Earrings", nil, oldEquipMap[*class.OldEquipmentSheet.Earring])
			classInfo.Equipment.Earring = &earrings
			rings := m.changeEquipToNewEquipType("Ring", nil, oldEquipMap[*class.OldEquipmentSheet.Ring])
			classInfo.Equipment.Ring = &rings
			cloak := m.changeEquipToNewEquipType("Cloak", nil, oldEquipMap[*class.OldEquipmentSheet.Cloak])
			classInfo.Equipment.Cloak = &cloak
			stockings := m.changeEquipToNewEquipType("Stockings", nil, oldEquipMap[*class.OldEquipmentSheet.Stockings])
			classInfo.Equipment.Stockings = &stockings
		}
		newClassMap[class.Name] = classInfo
		newClassMap[class.Name].OldEquipmentSheet = nil
		newClassMap[class.Name].CurrentWeapon = nil
	}
	copyUser.ClassMap = newClassMap
	_, err := m.users.UpdateDocument(copyUser.ID, copyUser)
	if err != nil {
		m.log.Errorf("error updating user:%v", err)
		return err
	}
	return nil
}

func (m *manage) changeEquipToNewEquipType(equipmentType string, currentWeapon *string, sheet *models.OldEquipmentSheet) models.Item {
	m.log.Debugf("equipmentType: %v", equipmentType)

	queryArgs := &[]models.QueryArg{}
	if equipmentType == "Weapon" {
		m.log.Debugf("currentWeapon: %v", sheet.LevelRequirement)
		m.log.Debugf("currentWeapon: %s", *currentWeapon)
		queryArgs = &[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: int(sheet.LevelRequirement),
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: *currentWeapon,
			},
		}
	} else {
		queryArgs = &[]models.QueryArg{
			{
				Path:  "levelRequirement",
				Op:    "==",
				Value: int(sheet.LevelRequirement),
			},
			{
				Path:  "type.weaponType",
				Op:    "==",
				Value: equipmentType,
			},
		}
	}
	items, err := m.item.QueryDocuments(queryArgs)
	if err != nil {
		m.log.Errorf("error querying for items: %v", err)
		panic("error encountered while getting items in batch")
	}
	m.log.Errorf("items: %v", items)
	return items[0]
}

func (m *manage) AddNewBoss(boss models.Monster) (*string, error) {
	id, err := m.boss.InsertDocument(&boss.Name, boss)
	if err != nil {
		m.log.Errorf("error inserting document for boss: %v", err)
		return nil, err
	}
	return id, nil
}

func (m *manage) AddNewItem(item *models.Item) (*string, error) {
	if !utils.ValidItemType(item.Type) {
		message := fmt.Sprintf("An invalid piece of equipment was entered")
		return &message, nil
	}
	id, err := m.item.InsertDocument(item)
	if err != nil {
		m.log.Errorf("error adding new item: %v", err)
		return nil, err
	}
	return id, nil
}

func (m *manage) ToggleExpEvent(expRate int) error {
	expMap := make(map[string]int)
	expMap["exp"] = expRate
	_, err := m.config.UpdateDocument(expMap, "exp")
	if err != nil {
		m.log.Errorf("error toggling exp flag: %v", err)
		return err
	}
	return nil
}

func (m *manage) IncreaseLevelCap(level int) (*[]models.Level, error) {
	levels := m.calculateExpTable(level)
	currentLevels, err := m.levels.QueryDocuments(globals.LEVELS, nil)
	if err != nil {
		m.log.Errorf("error retrieving current levels: %v", err)
		return nil, err
	}
	var addedLevels []models.Level
	for _, level := range levels {
		stringLevel := utils.ThirtyTwoBitIntToString(level.Value)
		m.log.Debugf("level.Val: %s", stringLevel)
		if currentLevels[stringLevel] == nil {
			insertedLevel, err := m.levels.InsertDocument(&stringLevel, level)
			if err != nil {
				m.log.Errorf("error inserting level: %v new levels: %v", level.Value, err)
				return nil, err
			}
			addedLevels = append(addedLevels, *insertedLevel)
		}
	}
	_, err = m.levels.UpdateDocument(globals.LEVELCAP, levels[len(levels)-1])
	if err != nil {
		m.log.Errorf("error updating level cap: %v", err)
		return nil, err
	}
	return &addedLevels, nil
}

func (m *manage) CreateExpTable(levels []models.Level) (*[]models.Level, error) {
	currentLevels, err := m.levels.QueryDocuments(globals.LEVELS, nil)
	if err != nil {
		m.log.Errorf("error retrieving current levels: %v", err)
		return nil, err
	}
	var addedLevels []models.Level
	for _, level := range levels {
		stringLevel := utils.ThirtyTwoBitIntToString(level.Value)
		m.log.Debugf("level.Val: %s", stringLevel)
		if currentLevels[stringLevel] == nil {
			insertedLevel, err := m.levels.InsertDocument(&stringLevel, level)
			if err != nil {
				m.log.Errorf("error inserting level: %v with error: %v", level.Value, err)
				return nil, err
			}
			addedLevels = append(addedLevels, *insertedLevel)
		} else {
			_, err := m.levels.UpdateDocument(stringLevel, level)
			if err != nil {
				m.log.Errorf("error updating level: %v with error: %v", level.Value, err)
				return nil, err
			}
			addedLevels = append(addedLevels, level)
		}
	}
	_, err = m.levels.UpdateDocument(globals.LEVELCAP, levels[len(levels)-1])
	if err != nil {
		m.log.Errorf("error updating level cap: %v", err)
		return nil, err
	}
	return &addedLevels, nil
}

func (m *manage) AddNewArea(area models.Area) (*string, error) {
	id, err := m.areas.InsertDocument(&area.ID, area)
	if err != nil {
		m.log.Errorf("error adding area: %v", err)
		return nil, err
	}
	return id, nil
}

//func (m *manage) BatchBossItems() (*string, error) {
//	items, err := m.item.QueryDocuments(&[]models.QueryArg{})
//	if err != nil {
//		m.log.Errorf("error getting items: %v", err)
//		return nil, err
//	}
//
//	m.item.UpdateDocument()
//}

func (m *manage) AddNewMonster(area *models.Area, monster models.Monster) (*string, error) {
	area.Monsters = append(area.Monsters, monster)
	time, err := m.areas.UpdateDocument(area.ID, area)
	if err != nil {
		m.log.Errorf("error updating area with new monster: %v", err)
		return nil, err
	}
	insertTime := time.String()
	return &insertTime, nil
}

func (m *manage) AddNewClass(class *models.JobClass) (*string, error) {
	id, err := m.classes.InsertDocument(&class.Name, class)
	if err != nil {
		m.log.Errorf("error adding class: %v", err)
		return nil, err
	}
	return id, nil
}

func (m *manage) AddNewUser(user models.User, weapon string) (*string, *string, error) {
	_, err := m.users.ReadDocument(user.ID)
	if err == nil {
		s := fmt.Sprintf("You already have an account affiliated with this bot.")
		return nil, &s, nil
	}
	classExist, err := m.classes.ReadDocument(strings.Title(strings.ToLower(user.CurrentClass)))
	if err != nil {
		m.log.Errorf("error getting current class: %v", err)
		s := fmt.Sprintf("The %s class does not exist.  Please select a valid class with a valid weapon", user.CurrentClass)
		return nil, &s, nil
	}
	if classExist.ClassRequirement != nil {
		message := fmt.Sprintf("Please selected a starting class.  %s is not a starting class.", classExist.Name)
		return nil, &message, nil
	}

	cleanedWeaponName := strings.Title(strings.ToLower(weapon))
	for _, weapon := range classExist.Weapons {
		if cleanedWeaponName == weapon.Name {
			newUser := m.generateNewUser(user, classExist.Name, cleanedWeaponName)
			id, err := m.users.InsertDocument(&newUser.ID, newUser)
			if err != nil {
				m.log.Errorf("error adding user: %v", err)
				return nil, nil, err
			}
			return id, nil, nil
		}

	}
	message := fmt.Sprintf("weapon: %s does not exist or cannot be equipped by your current class!", cleanedWeaponName)
	return nil, &message, nil
}

func (m *manage) AddNewEquipmentSheet(equipment models.OldEquipmentSheet) (*string, error) {
	var weaponMap = make(map[string]string)
	for _, weapon := range equipment.WeaponList {
		weaponMap[weapon.Type] = weapon.Name
	}
	equipment.WeaponMap = weaponMap
	id, err := m.equipment.InsertDocument(&equipment.ID, equipment)
	if err != nil {
		m.log.Errorf("error adding new equipment sheet")
		return nil, err
	}
	return id, err
}

func (m *manage) calculateExpTable(level int) []models.Level {
	var levels []models.Level
	for i := 1; i <= level; i++ {
		if i == 1 {
			levels = append(levels, models.Level{Value: 1, Exp: 0})
		} else {
			levels = append(levels, models.Level{Value: int32(i), Exp: int64(m.calculateExpForLevel(i))})
		}
	}
	return levels
}

func (m *manage) calculateExpForLevel(level int) int {
	// 50 * (level^2) - (50 * level)
	return 50*(int(math.Pow(float64(level), float64(2)))) - (50 * level)
}

func (m *manage) generateNewUser(user models.User, class, weapon string) models.User {
	startingWeapon, err := m.item.QueryForDocument(&[]models.QueryArg{
		{
			Path:  "levelRequirement",
			Op:    "==",
			Value: 1,
		},
		{
			Path:  "type.weaponType",
			Op:    "==",
			Value: weapon,
		},
	})
	if err != nil {
		panic("error getting weapons")
	}
	top, err := m.item.QueryForDocument(&[]models.QueryArg{
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
	bottom, err := m.item.QueryForDocument(&[]models.QueryArg{
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
	headpiece, err := m.item.QueryForDocument(&[]models.QueryArg{
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
	gloves, err := m.item.QueryForDocument(&[]models.QueryArg{
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
	boots, err := m.item.QueryForDocument(&[]models.QueryArg{
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
	newClass := make(map[string]*models.ClassInfo)
	newClass[strings.Title(strings.ToLower(user.CurrentClass))] = &models.ClassInfo{
		Name:  class,
		Level: 1,
		Exp:   0,
		Equipment: models.Equipment{
			Weapon:    *startingWeapon,
			Top:       *top,
			Bottom:    *bottom,
			Headpiece: *headpiece,
			Glove:     *gloves,
			Shoes:     *boots,
		},
	}
	beginnerEly := int64(0)
	return models.User{
		ID:           user.ID,
		CurrentClass: strings.Title(strings.ToLower(user.CurrentClass)),
		ClassMap:     newClass,
		Ely:          &beginnerEly,
		Name:         user.Name,
	}
}
