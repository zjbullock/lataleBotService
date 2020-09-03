package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/models"
	"math"
	"math/rand"
)

type Damage interface {
	DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool) (string, int)
	DetermineBossDamage(randGenerator *rand.Rand, user models.UserBlob, boss *models.Monster, bossSkill *models.BossSkill) (*models.UserBlob, string, int)
}

type damage struct {
	log loggo.Logger
}

func NewDamageService(log loggo.Logger) Damage {
	return &damage{
		log: log,
	}
}

func (d *damage) DetermineBossDamage(randGenerator *rand.Rand, user models.UserBlob, boss *models.Monster, bossSkill *models.BossSkill) (*models.UserBlob, string, int) {
	evasionChance := randGenerator.Float64()
	accuracy := boss.Stats.Accuracy
	if accuracy > 1.0 {
		accuracy = 1.0
	}
	evasion := user.StatModifier.Evasion
	if evasion > 1.0 {
		evasion = 1.0
	}
	if evasionChance > accuracy-evasion {
		return &user, fmt.Sprintf("__**%s**__ successfully ***EVADED*** __**%s**__'s attack!", user.User.Name, boss.Name), 0
	}
	damageLog := ""
	damage := float64(rand.Intn(int(boss.Stats.MaxDPS)-int(boss.Stats.MinDPS))) + boss.Stats.MinDPS

	updatedUser := &user
	if bossSkill != nil {
		damage = damage * bossSkill.SkillDamageModifier
		damageLog += fmt.Sprintf(bossSkill.Quote+" with the skill ***%s*** ", user.User.Name, bossSkill.Name)
		if bossSkill.CrowdControl != nil && (updatedUser.CrowdControlled == nil || *updatedUser.CrowdControlled == 0) {
			updatedUser.CrowdControlled = bossSkill.CrowdControl
			updatedUser.CrowdControlStatus = bossSkill.CrowdControlStatus
		}
	} else {
		damageLog += fmt.Sprintf("__**%s**__ hit __**%s**__ ", boss.Name, user.User.Name)
	}
	roundedDamage := ((int(damage) - int(user.StatModifier.Defense)) + int(math.Abs(damage-user.StatModifier.Defense))) / 2
	criticalChance := randGenerator.Float64()
	damageLog += fmt.Sprintf("for ")
	if boss.Stats.CriticalRate != 0.0 && criticalChance <= boss.Stats.CriticalRate {
		roundedDamage = int(float64(roundedDamage) * boss.Stats.CriticalDamageModifier)
		damageLog += fmt.Sprintf("**%v** ***CRITICAL*** damage!", roundedDamage)
	} else {
		damageLog += fmt.Sprintf("**%v** damage!", roundedDamage)
	}
	return updatedUser, damageLog, roundedDamage
}

func (d *damage) DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool) (string, int) {

	evasionChance := randGenerator.Float64()
	accuracy := attacker.Accuracy
	if accuracy > 1.0 {
		accuracy = 1.0
	}
	evasion := defender.Evasion
	if evasion > 1.0 {
		evasion = 1.0
	}
	if evasionChance > accuracy-evasion {
		return fmt.Sprintf("__**%s**__ successfully ***EVADED*** __**%s**__'s attack!", defenderName, attackerName), 0
	}
	theMonster := " "
	if !boss && weapon != nil {
		theMonster = " the "
	}
	damageLog := fmt.Sprintf("__**%s**__ hit%s__**%s**__ ", attackerName, theMonster, defenderName)
	damage := float64(rand.Intn(int(attacker.MaxDPS)-int(attacker.MinDPS))) + attacker.MinDPS
	skillChance := randGenerator.Float64()
	if attacker.SkillProcRate != 0.0 && skillChance <= attacker.SkillProcRate {
		skillName, damageMod := d.getSkill(randGenerator, *weapon, *class, int(*userLevel))
		damage = damage * 1.25 * damageMod
		damageLog += fmt.Sprintf("with the skill ***%s*** ", skillName)
	}
	roundedDamage := ((int(damage) - int(defender.Defense)) + int(math.Abs(damage-defender.Defense))) / 2

	criticalChance := randGenerator.Float64()
	damageLog += fmt.Sprintf("for ")
	if attacker.CriticalRate != 0.0 && criticalChance <= attacker.CriticalRate {
		roundedDamage = int(float64(roundedDamage) * attacker.CriticalDamageModifier)
		damageLog += fmt.Sprintf("**%v** ***CRITICAL*** damage!", roundedDamage)
	} else {
		damageLog += fmt.Sprintf("**%v** damage!", roundedDamage)
	}

	return damageLog, roundedDamage
}

func (d *damage) getSkill(randGenerator *rand.Rand, currentWeapon string, jobClass models.JobClass, userLevel int) (string, float64) {
	skill := ""
	damageMod := 1.0
	for _, weapon := range jobClass.Weapons {
		if currentWeapon == weapon.Name {
			skillTier := randGenerator.Intn(100)
			tier := 0
			if userLevel/20 >= 2 && skillTier >= 81 {
				tier = 2
				if userLevel/20 >= 4 && jobClass.Tier >= 2 {
					tierThree := randGenerator.Intn(100)
					if tierThree > 50 {
						tier = 3
						if userLevel/20 >= 6 && jobClass.Tier >= 3 {
							tierFour := randGenerator.Intn(100)
							if tierFour > 50 {
								tier = 4
							}
						}
					}
				}
			} else if userLevel/20 >= 1 && skillTier <= 70 && skillTier >= 51 {
				tier = 1
			}
			skill = weapon.Skills[tier]
			damageMod = ((float64(tier) / 10.0 * 2.0) + 1.0) * jobClass.Stats.SkillDamageModifier
		}
	}
	return skill, damageMod
}
