package services

import (
	"github.com/juju/loggo"
	"lataleBotService/models"
	"lataleBotService/repositories"
)

type Adventure interface {
	GetBaseStat(id string) (*models.StatModifier, *string, error)
	GetArea(id string) (*models.Area, error)
}

type adventure struct {
	areas   repositories.AreasRepository
	classes repositories.ClassRepository
	users   repositories.UserRepository
	log     loggo.Logger
}

func NewAdventureService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, log loggo.Logger) Adventure {
	return &adventure{
		areas:   areas,
		classes: classes,
		users:   users,
		log:     log,
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

func (a *adventure) calculateBaseStat(level float64, class models.StatModifier) models.StatModifier {
	levelModifier := float64((level / 100) + 1)
	return models.StatModifier{
		DPS:                    getDynamicStat(10, levelModifier, level, class.DPS),
		Defense:                getDynamicStat(10, levelModifier, level, class.Defense),
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
	return baseStat * levelModifier * level * statModifier
}

func getStaticStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * levelModifier * statModifier
}
