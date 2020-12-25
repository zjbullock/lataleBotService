package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"math"
	"math/rand"
)

type Damage interface {
	DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool) (string, int, *models.CrowdControlTrait)
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
func (d *damage) DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool) (string, int, *models.CrowdControlTrait) {
	damageLog := ""
	hitCount := 1
	var statusAilment models.CrowdControlTrait

	if weapon != nil && class != nil && class.Trait != nil && class.Trait.Type == globals.ATTACKTRAIT {
		traitChance := randGenerator.Float64()
		if traitChance <= *class.Trait.ActivationRate {
			damageLog += fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!", attackerName, class.Trait.Name)
			if class.Trait.Battle != nil {
				hitCount = int(class.Trait.Battle.HitCounter)
				newStats := attacker.AddBuffStats(class.Trait.Battle.Buff.StatModifier)
				attacker.AddStatModifier(newStats)
			}
			if class.Trait.CrowdControl != nil {
				statusAilment = *class.Trait.CrowdControl
			}
		}
	}
	d.log.Infof("Skill damage Modifier: %v", attacker.SkillDamageModifier)
	roundedDamage := 0
	totalDamage := 0
	for i := 0; i < hitCount; i++ {
		if damageLog != "" {
			damageLog += "\n"
		}
		evasionChance := randGenerator.Float64()
		accuracy := attacker.Accuracy
		evasion := defender.Evasion
		if evasion > 1.0 {
			evasion = 1.0
		}
		if evasion > 0.0 && evasionChance > accuracy-evasion {
			return fmt.Sprintf("__**%s**__ successfully ***EVADED*** __**%s**__'s attack!", defenderName, attackerName), 0, nil
		}

		if weapon == nil && class != nil && class.Trait != nil && class.Trait.Type == globals.GUARDTRAIT {
			skillActivation := randGenerator.Float64()
			if skillActivation < *class.Trait.ActivationRate {
				return fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!  Damage nullified!", defenderName, class.Trait.Name), 0, nil
			}
		}
		theMonster := " "
		if !boss && weapon != nil {
			theMonster = " the "
		}
		damageLog += fmt.Sprintf("__**%s**__ hit%s__**%s**__ ", attackerName, theMonster, defenderName)
		damage := float64(rand.Intn(int(attacker.MaxDPS)-int(attacker.MinDPS))) + attacker.MinDPS
		skillChance := randGenerator.Float64()
		if attacker.SkillProcRate != 0.0 && skillChance <= attacker.SkillProcRate {
			skillName, damageMod := d.getSkill(randGenerator, *weapon, *class, int(*userLevel), attacker.SkillDamageModifier)
			damage = damage * 1.25 * damageMod
			damageLog += fmt.Sprintf("with the skill ***%s*** ", skillName)
		}
		defenderDefense := defender.Defense - (defender.Defense * attacker.TargetDefenseDecrease)
		roundedDamage = ((int(damage) - int(defenderDefense)) + int(math.Abs(damage-defenderDefense))) / 2

		criticalChance := randGenerator.Float64()
		damageLog += fmt.Sprintf("for ")
		if attacker.CriticalRate != 0.0 && criticalChance <= attacker.CriticalRate {
			roundedDamage = int(float64(roundedDamage) * attacker.CriticalDamageModifier)
			damageLog += fmt.Sprintf("**%v** ***CRITICAL*** damage!", roundedDamage)
		} else {
			damageLog += fmt.Sprintf("**%v** damage!", roundedDamage)
		}
		totalDamage += roundedDamage
	}

	return damageLog, totalDamage, &statusAilment
}

func (d *damage) DetermineBossDamage(randGenerator *rand.Rand, user models.UserBlob, boss *models.Monster, bossSkill *models.BossSkill) (*models.UserBlob, string, int) {
	evasionChance := randGenerator.Float64()
	accuracy := boss.Stats.Accuracy
	evasion := user.BattleStats.Evasion
	if evasion > 1.0 {
		evasion = 1.0
	}
	if evasion > 0.0 && evasionChance > accuracy-evasion {
		return &user, fmt.Sprintf("__**%s**__ successfully ***EVADED*** __**%s**__'s attack!", user.User.Name, boss.Name), 0
	}

	if user.JobClass.Trait != nil && user.JobClass.Trait.Type == globals.GUARDTRAIT {
		skillActivation := randGenerator.Float64()
		if skillActivation < *user.JobClass.Trait.ActivationRate {
			return &user, fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!  Damage nullified!", user.User.Name, user.JobClass.Trait.Name), 0
		}
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
	roundedDamage := ((int(damage) - int(user.BattleStats.Defense)) + int(math.Abs(damage-user.BattleStats.Defense))) / 2
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

func (d *damage) getSkill(randGenerator *rand.Rand, currentWeapon string, jobClass models.JobClass, userLevel int, skillDamageModifier float64) (string, float64) {
	skill := ""
	damageMod := 1.0
	for _, weapon := range jobClass.Weapons {
		if currentWeapon == weapon.Name {
			skillTier := randGenerator.Float64()
			tier := 0
			if userLevel/20 >= 2 && skillTier > 0.81 {
				tier = 2
				if userLevel/20 >= 4 && jobClass.Tier >= 2 {
					tierThree := randGenerator.Float64()
					if tierThree >= 0.50 {
						tier = 3
						if userLevel >= 125 && jobClass.Tier >= 3 {
							tierFour := randGenerator.Float64()
							if tierFour >= 0.75 {
								tier = 4
							}
						}
					}
				}
			} else if userLevel/20 >= 1 && skillTier <= 0.81 && skillTier >= 0.51 {
				tier = 1
			}
			skill = weapon.Skills[tier]
			damageMod = ((float64(tier) / 10.0 * 2.0) + 1.0) * skillDamageModifier
		}
	}
	return skill, damageMod
}
