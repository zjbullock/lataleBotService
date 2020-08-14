package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/repositories"
	"lataleBotService/utils"
	"math"
	"strings"
)

type Manage interface {
	AddNewUser(user models.User) (*string, *string, error)
	AddNewClass(class models.JobClass) (*string, error)
	AddNewArea(area models.Area) (*string, error)
	AddNewMonster(area string, monster models.Monster) (*string, error)
	IncreaseLevelCap(level int) (*[]models.Level, error)
	AddNewEquipmentSheet(equipment models.EquipmentSheet) (*string, error)
}

type manage struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	levels    repositories.LevelRepository
	users     repositories.UserRepository
	equipment repositories.EquipmentRepository
	log       loggo.Logger
}

func NewManageService(areas repositories.AreasRepository, levels repositories.LevelRepository, classes repositories.ClassRepository, users repositories.UserRepository, equip repositories.EquipmentRepository, log loggo.Logger) Manage {
	return &manage{
		areas:     areas,
		classes:   classes,
		levels:    levels,
		users:     users,
		equipment: equip,
		log:       log,
	}
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
		stringLevel := utils.String(level.Value)
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

func (m *manage) AddNewArea(area models.Area) (*string, error) {
	id, err := m.areas.InsertDocument(&area.Name, area)
	if err != nil {
		m.log.Errorf("error adding area: %v", err)
		return nil, err
	}
	return id, nil
}

func (m *manage) AddNewMonster(area string, monster models.Monster) (*string, error) {
	panic("implement me!")
}

func (m *manage) AddNewClass(class models.JobClass) (*string, error) {
	id, err := m.classes.InsertDocument(&class.Name, class)
	if err != nil {
		m.log.Errorf("error adding class: %v", err)
		return nil, err
	}
	return id, nil
}

func (m *manage) AddNewUser(user models.User) (*string, *string, error) {
	_, err := m.users.ReadDocument(user.Name)
	if err == nil {
		s := fmt.Sprintf("user already exists")
		return nil, &s, nil
	}
	classExist, err := m.classes.ReadDocument(strings.Title(strings.ToLower(user.CurrentClass)))
	if err != nil {
		m.log.Errorf("error getting current class: %v", err)
		s := fmt.Sprintf("Class %s does not exist, you big dumb.", user.CurrentClass)
		return nil, &s, nil
	}

	cleanedWeaponName := strings.Title(strings.ToLower(user.CurrentWeapon))
	for _, weapon := range classExist.Weapons {
		if cleanedWeaponName == weapon.Name {
			user.CurrentWeapon = cleanedWeaponName
			newUser := m.generateNewUser(user, classExist.Name)
			id, err := m.users.InsertDocument(&newUser.Name, newUser)
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

func (m *manage) AddNewEquipmentSheet(equipment models.EquipmentSheet) (*string, error) {
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
			levels = append(levels, models.Level{Value: int32(i), Exp: int32(m.calculateExpForLevel(i))})
		}
	}
	return levels
}

func (m *manage) calculateExpForLevel(level int) int {
	return 50*(int(math.Pow(float64(level), float64(2)))) - (50 * level)
}

func (m *manage) generateNewUser(user models.User, class string) models.User {
	newClass := make(map[string]models.ClassInfo)
	newClass[strings.Title(strings.ToLower(user.CurrentClass))] = models.ClassInfo{
		Name:  class,
		Level: 1,
		Exp:   0,
		Equipment: models.Equipment{
			Weapon: 0,
			Body:   0,
			Glove:  0,
			Shoes:  0,
		},
	}
	ely := int32(0)
	level := int32(1)
	return models.User{
		CurrentWeapon: user.CurrentWeapon,
		CurrentClass:  strings.Title(strings.ToLower(user.CurrentClass)),
		ClassMap:      newClass,
		Ely:           &ely,
		Name:          user.Name,
		CurrentLevel:  &level,
	}
}
