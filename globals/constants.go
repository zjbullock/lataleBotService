package globals

const (
	CLASSES   = "classes"
	USERS     = "users"
	AREAS     = "areas"
	LEVELS    = "levels"
	CONFIG    = "config"
	PARTY     = "parties"
	ITEM      = "items"
	SETBONUS  = "setBonus"
	CITY      = "cities"
	LEVELCAP  = "levelCap"
	EQUIPMENT = "equipment"
	BOSSES    = "bosses"
)

type TraitType string

const (
	REACTIVETRAIT        TraitType = "REACTIVETRAIT"
	GUARDTRAIT           TraitType = "GUARDTRAIT"
	HPPERCENTTRAIT       TraitType = "HPPERCENTTRAIT"
	DEATHTRAIT           TraitType = "DEATHTRAIT"
	ATTACKTRAIT          TraitType = "ATTACKTRAIT"
	BATTLESTARTTRAIT     TraitType = "BATTLESTARTTRAIT"
	SUMMONTRAIT          TraitType = "SUMMONTRAIT"
	AFTERATTACKTRAIT     TraitType = "AFTERATTACKTRAIT"
	HEALTRAIT            TraitType = "HEALTRAIT"
	DEFENDINGCHANCETRAIT TraitType = "DEFENDINGCHANCETRAIT"
)
