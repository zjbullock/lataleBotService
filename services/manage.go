package services

import (
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/repositories"
	"lataleBotService/utils"
	"math"
)

type Manage interface {
	AddNewUser(user models.User) (*string, error)
	AddNewClass(class models.JobClass) (*string, error)
	AddNewArea(area models.Area) (*string, error)
	AddNewMonster(area string, monster models.Monster) (*string, error)
	IncreaseLevelCap(level int) (*[]models.Level, error)
}

type manage struct {
	areas   repositories.AreasRepository
	classes repositories.ClassRepository
	levels  repositories.LevelRepository
	users   repositories.UserRepository
	log     loggo.Logger
}

func NewManageService(areas repositories.AreasRepository, levels repositories.LevelRepository, classes repositories.ClassRepository, users repositories.UserRepository, log loggo.Logger) Manage {
	return &manage{
		areas:   areas,
		classes: classes,
		levels:  levels,
		users:   users,
		log:     log,
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

func (m *manage) AddNewUser(user models.User) (*string, error) {
	id, err := m.users.InsertDocument(&user.Name, user)
	if err != nil {
		m.log.Errorf("error adding user: %v", err)
		return nil, err
	}
	return id, nil
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
