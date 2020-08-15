package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/models"
	"math"
	"math/rand"
)

type Damage interface {
	DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32) (string, int)
}

type damage struct {
	log loggo.Logger
}

func NewDamageService(log loggo.Logger) Damage {
	return &damage{
		log: log,
	}
}

func (d *damage) DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32) (string, int) {

	evasionChance := randGenerator.Float64()
	d.log.Debugf("%s Evasion Chance: %v", defenderName, evasionChance)
	if evasionChance > attacker.Accuracy-defender.Evasion {
		return fmt.Sprintf("%s successfully evaded %s's attack!", defenderName, attackerName), 0
	}
	damageLog := fmt.Sprintf("%s hit the %s ", attackerName, defenderName)
	damage := float64(rand.Intn(int(attacker.MaxDPS)-int(attacker.MinDPS))) + attacker.MinDPS
	criticalChance := randGenerator.Float64()
	d.log.Debugf("%s Critical Chance: %v", attackerName, criticalChance)
	if attacker.CriticalRate != 0.0 && criticalChance <= attacker.CriticalRate {
		damage = damage * attacker.CriticalDamageModifier
		damageLog += "CRITICALLY "
	}
	skillChance := randGenerator.Float64()
	d.log.Debugf("%s Skill Chance: %v", attackerName, skillChance)
	if attacker.SkillProcRate != 0.0 && skillChance <= attacker.SkillProcRate {
		skillName, damageMod := d.getSkill(randGenerator, *weapon, *class, int(*userLevel))
		damage = damage * 1.25 * damageMod
		damageLog += fmt.Sprintf("with the skill %s ", skillName)
	}
	roundedDamage := ((int(damage) - int(defender.Defense)) + int(math.Abs(damage-defender.Defense))) / 2
	damageLog += fmt.Sprintf("for %v damage!", roundedDamage)
	return damageLog, roundedDamage
}

func (d *damage) getSkill(randGenerator *rand.Rand, currentWeapon string, jobClass models.JobClass, userLevel int) (string, float64) {
	skill := ""
	damageMod := 1.0
	for _, weapon := range jobClass.Weapons {
		if currentWeapon == weapon.Name {
			skillTier := randGenerator.Intn(100)
			tier := 0
			if userLevel/20 >= 3 && skillTier >= 96 {
				tier = 2
			} else if userLevel/20 >= 2 && skillTier <= 95 && skillTier >= 60 {
				tier = 1
			}
			skill = weapon.Skills[tier]
			damageMod = (float64(tier) / 10.0 * 2.0) + 1.0
		}
	}
	return skill, damageMod
}
