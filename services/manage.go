package services

import (
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/repositories"
)

type Manage interface {
	AddNewUser(user models.User) (*string, error)
	AddNewClass(class models.JobClass) (*string, error)
	AddNewArea(area models.Area) (*string, error)
	AddNewMonster(area string, monster models.Monster) (*string, error)
	IncreaseLevelCap(level int) (*string, error)
}

type manage struct {
	areas   repositories.AreasRepository
	classes repositories.ClassRepository
	users   repositories.UserRepository
	log     loggo.Logger
}

func NewManageService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, log loggo.Logger) Manage {
	return &manage{
		areas:   areas,
		classes: classes,
		users:   users,
		log:     log,
	}
}

func (m *manage) IncreaseLevelCap(level int) (*string, error) {
	return nil, nil
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
	return nil, nil
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
