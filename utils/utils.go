package utils

import (
	"lataleBotService/models"
	"strconv"
)

func String(n int64) string {
	return strconv.FormatInt(n, 10)
}

func ThirtyTwoBitIntToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func ValidItemType(itemType models.ItemType) bool {
	if itemType.Type == "armor" {
		validArmorTypes := map[string]bool{
			"Headpiece": true,
			"Top":       true,
			"Bottom":    true,
			"Boots":     true,
			"Gloves":    true,
			"Bindi":     true,
			"Earrings":  true,
			"Ring":      true,
			"Glasses":   true,
			"Stockings": true,
			"Cloak":     true,
		}
		return validArmorTypes[*itemType.WeaponType]
	} else if itemType.Type == "weapon" {
		validWeaponTypes := map[string]bool{
			"Dagger":        true,
			"Crossbow":      true,
			"Bow":           true,
			"Doubleblades":  true,
			"Dualpistols":   true,
			"Greatsword":    true,
			"Guitar":        true,
			"Knuckles":      true,
			"Longsword":     true,
			"Mace":          true,
			"Mg":            true,
			"Orb":           true,
			"Spear":         true,
			"Staff":         true,
			"Spiralsword":   true,
			"Psionicblades": true,
			"Gauntlet":      true,
			"Psychichands":  true,
			"Battlestaff":   true,
			"Musicrod":      true,
			"Rogueknife":    true,
			"Gunblade":      true,
			"Guardianball":  true,
		}
		return validWeaponTypes[*itemType.WeaponType]
	} else if itemType.Type == "consumable" {

	} else if itemType.Type == "event" {

	}
	return false
}
