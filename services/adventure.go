package services

import (
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
	GetBaseStat(id string) (*models.StatModifier, *string, error)
	GetJobList() (*[]models.JobClass, error)
	GetAdventure(areaId, userId string) (*[]string, *string, error)
	GetJobClassDescription(id string) (*models.JobClass, error)
	GetArea(id string) (*models.Area, *string, error)
	GetAreas() (*[]models.Area, error)
	GetUserInfo(id string) (*models.User, *string, error)
}

type adventure struct {
	areas     repositories.AreasRepository
	classes   repositories.ClassRepository
	users     repositories.UserRepository
	equipment repositories.EquipmentRepository
	damage    Damage
	log       loggo.Logger
}

func NewAdventureService(areas repositories.AreasRepository, classes repositories.ClassRepository, users repositories.UserRepository, equips repositories.EquipmentRepository, log loggo.Logger) Adventure {
	return &adventure{
		areas:     areas,
		classes:   classes,
		users:     users,
		equipment: equips,
		damage:    NewDamageService(log),
		log:       log,
	}
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

		a.log.Errorf("classEquipmentList: %v", classEquipmentList)
		classInfo := user.ClassMap[class.Name]
		classInfo.Equipment.EquipmentNames = classEquipmentList
		user.ClassMap[class.Name] = classInfo
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
		message := "Could not find an area with that name.  Please be sure to use the correct name."
		return nil, &message, err
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
	monsters := a.determineMonsterRarity(monsterMap)
	if monsters == nil {
		afraid := fmt.Sprintf("The monsters in the %s are too afraid of fighting %s", areaId, userId)
		return &[]string{afraid}, nil, nil
	}
	monster := a.determineMonster(*monsters)
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
	adventureLog := a.createAdventureLog(*classInfo, user, *currentStats, monster)
	return &adventureLog, nil, nil
}

func (a *adventure) createAdventureLog(classInfo models.JobClass, user *models.User, userStats models.StatModifier, monster models.Monster) []string {
	var adventureLog []string
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
	adventureLog = append(adventureLog, fmt.Sprintf("%s has encountered a %s%s", user.Name, monster.Name, rankExclamation))
	userLevel := user.ClassMap[user.CurrentClass].Level
	userWeapon := user.ClassMap[user.CurrentClass].CurrentWeapon
	for currentHP != 0 && monsterHP != 0 {
		userLog, damage := a.damage.DetermineHit(randGenerator, user.Name, monster.Name, userStats, monster.Stats, &userWeapon, &classInfo, &userLevel)
		monsterHP = ((int(monsterHP) - int(damage)) + int(math.Abs(float64(monsterHP-damage)))) / 2
		adventureLog = append(adventureLog, userLog)
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s HP: %v/%v", monster.Name, monsterHP, monsterMaxHp))
		if monsterHP <= 0 {
			adventureLog = append(adventureLog, fmt.Sprintf("**%s has successfully defeated the %s!**", user.Name, monster.Name))
			battleWin = true
			break
		}
		monsterLog, damage := a.damage.DetermineHit(randGenerator, monster.Name, user.Name, monster.Stats, userStats, nil, nil, nil)
		currentHP = ((int(currentHP) - int(damage)) + int(math.Abs(float64(currentHP-damage)))) / 2
		adventureLog = append(adventureLog, monsterLog)
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s HP: %v/%v", user.Name, currentHP, userMaxHP))
		if currentHP <= 0 {
			adventureLog = append(adventureLog, fmt.Sprintf("**%s was killed by %s!**", user.Name, monster.Name))
			break
		}
		userHeal := int(userStats.HP * userStats.Recovery)
		if userHeal+currentHP > int(userStats.HP) {
			currentHP = int(userStats.HP)
		} else {
			currentHP += userHeal
		}
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ ***HEALED*** for %v HP.", user.Name, userHeal))
		adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s HP: %v/%v!", user.Name, currentHP, userMaxHP))
		if monster.Stats.Recovery > 0.0 {
			monsterHeal := int(monster.Stats.HP * monster.Stats.Recovery)
			if monsterHeal+monsterHP > int(monster.Stats.HP) {
				monsterHP = int(monster.Stats.HP)
			} else {
				monsterHP += monsterHeal
			}
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__ **HEALED** for %v HP.", monster.Name, monsterHeal))
			adventureLog = append(adventureLog, fmt.Sprintf("__**%s**__'s HP: %v/%v!", monster.Name, monsterHP, monsterMaxHp))
		}

	}
	if battleWin {
		//a.users.UpdateDocument(user.ID)
	}
	return adventureLog
}

func (a *adventure) determineMonsterRarity(monsterMap map[string]*[]models.Monster) *[]models.Monster {
	randSource := rand.NewSource(time.Now().UnixNano())
	rarityGenerator := rand.New(randSource)
	rarityPercent := rarityGenerator.Intn(100) + 1
	if monsterMap["3"] != nil && rarityPercent >= 96 {
		return monsterMap["3"]
	} else if monsterMap["2"] != nil && rarityPercent <= 95 && rarityPercent >= 60 {
		return monsterMap["2"]
	}
	return monsterMap["1"]
}

func (a *adventure) determineMonster(monsters []models.Monster) models.Monster {
	randSource := rand.NewSource(time.Now().UnixNano())
	monsterSelection := rand.New(randSource)
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
	return models.StatModifier{
		MaxDPS:                 getDynamicStat(20, levelModifier, level, class.MaxDPS) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].WeaponDPS,
		MinDPS:                 getDynamicStat(20, levelModifier, level, class.MinDPS) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Weapon)].WeaponDPS,
		Defense:                getDynamicStat(15, levelModifier, level, class.Defense) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Body)].ArmorDefense,
		HP:                     getDynamicStat(100, levelModifier, level, class.HP),
		Recovery:               getStaticStat(0.05, levelModifier, class.Recovery),
		CriticalDamageModifier: getStaticStat(1.5, levelModifier, class.CriticalDamageModifier),
		CriticalRate:           getStaticStat(0.05, levelModifier, class.CriticalRate),
		SkillProcRate:          getStaticStat(0.25, levelModifier, class.SkillProcRate),
		Evasion:                getStaticStat(0.05, levelModifier, class.Evasion) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Shoes)].ShoeEvasion,
		Accuracy:               getStaticStat(0.95, levelModifier, class.Accuracy) + equipmentMap[strconv.Itoa(user.ClassMap[user.CurrentClass].Equipment.Glove)].GloveAccuracy,
	}
}

func getDynamicStat(baseStat, levelModifier, level, statModifier float64) float64 {
	return baseStat * statModifier * math.Pow(levelModifier, 7)
}

func getStaticStat(baseStat, levelModifier, statModifier float64) float64 {
	return baseStat * levelModifier * statModifier
}
