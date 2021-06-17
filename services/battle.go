package services

import (
	"fmt"
	"github.com/juju/loggo"
	"lataleBotService/globals"
	"lataleBotService/models"
	"math"
	"math/rand"
	"strconv"
)

type Battle interface {
	DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool, hitCount int) (string, int, *models.CrowdControlTrait)
	DetermineBossDamage(randGenerator *rand.Rand, user models.UserBlob, boss *models.Monster, bossSkill *models.BossSkill) (*models.UserBlob, string, int)
	DetermineTraitActivations(users []*models.UserBlob, adventureLog []string, buffType globals.TraitType, start bool) ([]*models.UserBlob, string)
	DecreaseUserBuffDuration(user *models.UserBlob) (string, *models.UserBlob)
	DecreaseUserDebuffDuration(user *models.UserBlob) (string, *models.UserBlob)
	DecreaseMonsterDebuffDuration(monster *models.MonsterBlob) (string, *models.MonsterBlob)
	InflictStatusAilmentMonster(monster *models.MonsterBlob, statusAilment models.CrowdControlTrait) (string, *models.MonsterBlob)
	GenerateSummons(user string, summons []models.Summons, summonTrait models.SummonTrait, stats models.StatModifier) (string, []models.Summons)
	DecreaseSummonDuration(user string, summons []models.Summons) (string, []models.Summons)
}

type battle struct {
	log loggo.Logger
}

func NewBattleService(log loggo.Logger) Battle {
	return &battle{
		log: log,
	}
}

func (b *battle) DetermineHit(randGenerator *rand.Rand, attackerName, defenderName string, attacker, defender models.StatModifier, weapon *string, class *models.JobClass, userLevel *int32, boss bool, hitCount int) (string, int, *models.CrowdControlTrait) {
	damageLog := ""
	var statusAilment *models.CrowdControlTrait

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
				statusAilment = class.Trait.CrowdControl
			}
		}
	}
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
			skillName, damageMod := b.getSkill(randGenerator, *weapon, *class, int(*userLevel), attacker.SkillDamageModifier)
			damage = damage * 1.25 * damageMod
			damageLog += fmt.Sprintf("with the skill ***%s*** ", skillName)
		}

		defenderDefense := defender.Defense - (defender.Defense * (attacker.TargetDefenseDecrease - defender.DamageMitigation))
		roundedDamage = ((int(damage) - int(defenderDefense)) + int(math.Abs(damage-defenderDefense))) / 2

		criticalChance := randGenerator.Float64()
		damageLog += fmt.Sprintf("for ")
		if attacker.CriticalRate != 0.0 && criticalChance <= attacker.CriticalRate {
			roundedDamage = int(float64(roundedDamage) * attacker.CriticalDamageModifier)
			damageLog += fmt.Sprintf("**%v** ***CRITICAL*** Damage!", roundedDamage)
		} else {
			damageLog += fmt.Sprintf("**%v** Damage!", roundedDamage)
		}
		totalDamage += roundedDamage
	}

	return damageLog, totalDamage, statusAilment
}

func (b *battle) DetermineBossDamage(randGenerator *rand.Rand, user models.UserBlob, boss *models.Monster, bossSkill *models.BossSkill) (*models.UserBlob, string, int) {
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
		if bossSkill.CrowdControl != nil {
			var debuff *models.StatModifier
			if bossSkill.Debuff != nil {
				diffStats := updatedUser.BattleStats.AddBuffStats(*bossSkill.Debuff)
				updatedUser.BattleStats.SubtractStatModifier(diffStats)
				debuff = &diffStats
			}
			updatedUser.Debuffs[*bossSkill.CrowdControlStatus] = models.CrowdControlTrait{
				Type:             *bossSkill.CrowdControlStatus,
				Debuff:           debuff,
				Bind:             bossSkill.Bind,
				CrowdControlTime: *bossSkill.CrowdControl,
			}

		}
	} else {
		damageLog += fmt.Sprintf("__**%s**__ hit __**%s**__ ", boss.Name, user.User.Name)
	}
	roundedDamage := ((int(damage) - int(user.BattleStats.Defense)) + int(math.Abs(damage-user.BattleStats.Defense))) / 2
	criticalChance := randGenerator.Float64()
	damageLog += fmt.Sprintf("for ")
	if boss.Stats.CriticalRate != 0.0 && criticalChance <= boss.Stats.CriticalRate {
		roundedDamage = int(float64(roundedDamage) * boss.Stats.CriticalDamageModifier)
		damageLog += fmt.Sprintf("**%v** ***CRITICAL*** Damage!", roundedDamage)
	} else {
		damageLog += fmt.Sprintf("**%v** Damage!", roundedDamage)
	}
	if len(user.Summons) > 0 {
		var aliveSummons []models.Summons
		for i, summon := range updatedUser.Summons {
			if updatedUser.Summons[i].Duration == nil {
				damageToSummon := float64(rand.Intn(int(boss.Stats.MaxDPS)-int(boss.Stats.MinDPS))) + boss.Stats.MinDPS
				if bossSkill != nil {
					damageToSummon *= bossSkill.SkillDamageModifier
				}
				summonRoundedDamage := ((damageToSummon) - (summon.StatModifier.Defense)) + (math.Abs(damageToSummon-summon.StatModifier.Defense))/2
				updatedUser.Summons[i].StatModifier.HP -= summonRoundedDamage
				damageLog += fmt.Sprintf("\n __**%s**__ has also taken **%v** damage!\n", summon.Name+" "+strconv.Itoa(i+1), int(summonRoundedDamage))
				if updatedUser.Summons[i].StatModifier.HP > 0.0 {
					damageLog += fmt.Sprintf("__**%s**__'s remaining HP: **%v**\n", summon.Name+" "+strconv.Itoa(i+1), int(updatedUser.Summons[i].StatModifier.HP))
					aliveSummons = append(aliveSummons, updatedUser.Summons[i])
				} else {
					damageLog += fmt.Sprintf("__**%s**__'s has died!\n", summon.Name+" "+strconv.Itoa(i+1))
				}
			} else {
				aliveSummons = append(aliveSummons, updatedUser.Summons[i])
			}
		}
		updatedUser.Summons = aliveSummons
	}
	return updatedUser, damageLog, roundedDamage
}

func (b *battle) getSkill(randGenerator *rand.Rand, currentWeapon string, jobClass models.JobClass, userLevel int, skillDamageModifier float64) (string, float64) {
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

func (b *battle) DetermineTraitActivations(users []*models.UserBlob, adventureLog []string, buffType globals.TraitType, start bool) ([]*models.UserBlob, string) {
	buffString := ""
	buffedUsers := users
	for i, user := range users {

		if user.JobClass.Trait != nil && user.JobClass.Trait.Type == buffType &&
			(user.JobClass.Trait.ActivationRate == nil || (user.JobClass.Trait.ActivationRate != nil && *user.JobClass.Trait.ActivationRate >= rand.Float64())) &&
			(user.JobClass.Trait.HPTrigger == nil || (user.JobClass.Trait.HPTrigger != nil && float64(user.CurrentHP)/float64(user.MaxHP) <= *user.JobClass.Trait.HPTrigger)) {
			if user.JobClass.Trait.Battle != nil {
				if user.JobClass.Trait.Battle.AoE == true && len(users) > 1 {
					buffString += fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!\n", user.User.Name, user.JobClass.Trait.Name)

					for j, buffedUser := range buffedUsers {
						if _, ok := buffedUser.Buffs[user.JobClass.Trait.Name]; !ok || (user.JobClass.Trait.Battle.AoE == true) {
							diffStats := buffedUsers[j].BaseStats.AddBuffStats(user.JobClass.Trait.Battle.Buff.StatModifier)
							buffedUsers[j].BattleStats.AddStatModifier(diffStats)
							buffedUsers[j].Buffs[user.JobClass.Trait.Name] = models.Buff{
								StatModifier: diffStats,
								Duration:     user.JobClass.Trait.Battle.Buff.Duration,
							}
							if buffedUsers[j].MaxHP < int(buffedUsers[j].BattleStats.HP) && user.JobClass.Trait.Type == globals.BATTLESTARTTRAIT {
								buffedUsers[j].MaxHP = int(buffedUsers[j].BattleStats.HP)
								buffedUsers[j].CurrentHP = buffedUsers[j].MaxHP
							}
						}
					}
				} else {
					if _, ok := buffedUsers[i].Buffs[user.JobClass.Trait.Name]; !ok {
						buffString += fmt.Sprintf("__***%s***__ activated their trait, __**%s**__!\n", user.User.Name, user.JobClass.Trait.Name)
						diffStats := buffedUsers[i].BaseStats.AddBuffStats(user.JobClass.Trait.Battle.Buff.StatModifier)
						buffedUsers[i].BattleStats.AddStatModifier(diffStats)
						buffedUsers[i].Buffs[user.JobClass.Trait.Name] = models.Buff{
							StatModifier: diffStats,
							Duration:     user.JobClass.Trait.Battle.Buff.Duration,
						}
						if buffedUsers[i].MaxHP < int(buffedUsers[i].BattleStats.HP) && user.JobClass.Trait.Type == globals.BATTLESTARTTRAIT {
							buffedUsers[i].MaxHP = int(buffedUsers[i].BattleStats.HP)
							buffedUsers[i].CurrentHP = buffedUsers[i].MaxHP
						}
					}

				}
			}
			if user.JobClass.Trait.Summon != nil && start {
				summonLogs, newSummons := b.GenerateSummons(user.User.Name, user.Summons, *user.JobClass.Trait.Summon, *user.BattleStats)
				buffString += summonLogs
				buffedUsers[i].Summons = newSummons
			}
		}
	}
	return buffedUsers, buffString
}

func (b *battle) DecreaseUserBuffDuration(user *models.UserBlob) (string, *models.UserBlob) {
	removedBuffs := ""
	newBuffs := user.Buffs
	for name, _ := range user.Buffs {
		newBuff := user.Buffs[name]
		newDuration := *newBuff.Duration
		newDuration--
		newBuff.Duration = &newDuration
		user.Buffs[name] = newBuff
		if *user.Buffs[name].Duration <= int32(0) {
			user.BattleStats.SubtractStatModifier(user.Buffs[name].StatModifier)
			user.MaxHP = int(user.BattleStats.HP)
			removedBuffs += fmt.Sprintf("The effects of __**%s**__ have expired for __***%s***__\n", name, user.User.Name)
			delete(newBuffs, name)

		}
	}
	user.Buffs = newBuffs
	if removedBuffs == "" {
		return "", nil
	}
	return removedBuffs, user
}

func (b *battle) DecreaseUserDebuffDuration(user *models.UserBlob) (string, *models.UserBlob) {
	removedDebuffs := ""
	newDebuffs := user.Debuffs
	for name, debuff := range user.Debuffs {
		newCCTrait := user.Debuffs[name]
		newCCTrait.CrowdControlTime--
		user.Debuffs[name] = newCCTrait
		if user.Debuffs[name].CrowdControlTime <= int32(0) {
			if debuff.Debuff != nil {
				user.BattleStats.AddStatModifier(*user.Debuffs[name].Debuff)
			}
			user.MaxHP = int(user.BattleStats.HP)
			removedDebuffs += fmt.Sprintf("The effects of __**%s**__ have expired for __***%s***__\n", name, user.User.Name)

			delete(newDebuffs, name)

		} else if name != "poison" && name != "bleed" && name != "burn" {
			removedDebuffs += fmt.Sprintf("__**%s**__ has the status ailment of __**%s**__ for **%v turn(s)**.\n", user.User.Name, name, user.Debuffs[name].CrowdControlTime)
		}
	}
	user.Debuffs = newDebuffs
	if removedDebuffs == "" {
		return "", nil
	}
	return removedDebuffs, user
}

func (b *battle) DecreaseMonsterDebuffDuration(monster *models.MonsterBlob) (string, *models.MonsterBlob) {
	removedDebuffs := ""
	newDebuffs := monster.Debuffs
	for name, debuff := range monster.Debuffs {
		newCCTrait := monster.Debuffs[name]
		newCCTrait.CrowdControlTime--
		monster.Debuffs[name] = newCCTrait
		if monster.Debuffs[name].CrowdControlTime <= int32(0) {
			if debuff.Debuff != nil {
				monster.BattleStats.AddStatModifier(*monster.Debuffs[name].Debuff)
			}
			removedDebuffs += fmt.Sprintf("The effects of __**%s**__ have expired for __***%s***__\n", name, monster.Name)
			delete(newDebuffs, name)
		} else {
			removedDebuffs += fmt.Sprintf("__**%s**__ has the status ailment of __**%s**__ for **%v more turn(s)**.\n", monster.Name, name, monster.Debuffs[name].CrowdControlTime)
		}
	}
	monster.Debuffs = newDebuffs
	if removedDebuffs == "" {
		return "", nil
	}
	return removedDebuffs, monster
}

func (b *battle) InflictStatusAilmentMonster(monster *models.MonsterBlob, statusAilment models.CrowdControlTrait) (string, *models.MonsterBlob) {
	var debuffStats *models.StatModifier
	if _, ok := monster.Debuffs[statusAilment.Type]; !ok {
		if statusAilment.Debuff != nil {
			diffStats := monster.BattleStats.AddBuffStats(*statusAilment.Debuff)
			monster.BattleStats.SubtractStatModifier(diffStats)
			debuffStats = &diffStats
		}
		monster.Debuffs[statusAilment.Type] = models.CrowdControlTrait{
			Type:             statusAilment.Type,
			Debuff:           debuffStats,
			Bind:             statusAilment.Bind,
			CrowdControlTime: statusAilment.CrowdControlTime,
		}
		return fmt.Sprintf("__**%s**__ was inflicted with __**%s**__\n", monster.Name, statusAilment.Type), monster
	}
	return "", monster
}

func (b *battle) GenerateSummons(user string, summons []models.Summons, summonTrait models.SummonTrait, stats models.StatModifier) (string, []models.Summons) {
	newSummon := models.Summons{
		Name:         summonTrait.Summons.Name,
		StatModifier: b.getSummonStats(stats, summonTrait.Summons.StatModifier),
		Duration:     summonTrait.Summons.Duration,
	}
	for i := int32(0); i < summonTrait.Count; i++ {
		summons = append(summons, newSummon)
	}
	return fmt.Sprintf("__**%s**__ summoned ***%s!***", user, newSummon.Name), summons
}

func (b *battle) DecreaseSummonDuration(user string, summons []models.Summons) (string, []models.Summons) {
	summonLogs := ""
	var aliveSummons []models.Summons
	for i, _ := range summons {
		if summons[i].Duration != nil {
			newDuration := *summons[i].Duration
			newDuration--
			summons[i].Duration = &newDuration
			if *summons[i].Duration == 0 {
				summonLogs += fmt.Sprintf("__**%s %d**__ has faded out of existence!\n", summons[i].Name, i+1)
			} else {
				summonLogs += fmt.Sprintf("__**%s %d**__ will remain active for %d more turn(s)!\n", summons[i].Name, i+1, *summons[i].Duration)
				aliveSummons = append(aliveSummons, summons[i])
			}
		} else {
			aliveSummons = append(aliveSummons, summons[i])
		}
	}
	return summonLogs, aliveSummons
}

func (b *battle) getSummonStats(user, summon models.StatModifier) models.StatModifier {
	summon.AmplifyStatModifier(user)
	return summon
}
