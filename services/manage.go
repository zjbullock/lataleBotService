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
	AddNewUser(user models.User, weapon string) (*string, *string, error)
	AddNewClass(class *models.JobClass) (*string, error)
	AddNewArea(area models.Area) (*string, error)
	AddNewMonster(area *models.Area, monster models.Monster) (*string, error)
	IncreaseLevelCap(level int) (*[]models.Level, error)
	CreateExpTable(levels []models.Level) (*[]models.Level, error)
	ToggleExpEvent(expRate int) error
	AddNewEquipmentSheet(equipment models.EquipmentSheet) (*string, error)
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
	log       loggo.Logger
}

func NewManageService(areas repositories.AreasRepository, levels repositories.LevelRepository, classes repositories.ClassRepository, users repositories.UserRepository, equip repositories.EquipmentRepository, config repositories.ConfigRepository, boss repositories.BossRepository, log loggo.Logger) Manage {
	return &manage{
		areas:     areas,
		classes:   classes,
		levels:    levels,
		users:     users,
		equipment: equip,
		config:    config,
		boss:      boss,
		log:       log,
	}
}

func (m *manage) AddNewBoss(boss models.Monster) (*string, error) {
	id, err := m.boss.InsertDocument(&boss.Name, boss)
	if err != nil {
		m.log.Errorf("error inserting document for boss: %v", err)
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

func (m *manage) CreateExpTable(levels []models.Level) (*[]models.Level, error) {
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
			levels = append(levels, models.Level{Value: int64(i), Exp: int64(m.calculateExpForLevel(i))})
		}
	}
	return levels
}

func (m *manage) calculateExpForLevel(level int) int {
	// 50 * (level^2) - (50 * level)
	return 50*(int(math.Pow(float64(level), float64(2)))) - (50 * level)
}

func (m *manage) generateNewUser(user models.User, class, weapon string) models.User {
	newClass := make(map[string]*models.ClassInfo)
	newClass[strings.Title(strings.ToLower(user.CurrentClass))] = &models.ClassInfo{
		Name:          class,
		Level:         1,
		Exp:           0,
		CurrentWeapon: weapon,
		Equipment: models.Equipment{
			Weapon: 0,
			Body:   0,
			Glove:  0,
			Shoes:  0,
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
