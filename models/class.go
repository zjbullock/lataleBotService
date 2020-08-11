package models

type JobClass struct {
	Name             string       `json:"name" firestore:"name"`
	LevelRequirement int32        `json:"levelRequirement" firestore:"levelRequirement"`
	Weapons          []Weapon     `json:"weapons" firestore:"weapons"`
	Stats            StatModifier `json:"stats" firestore:"stats"`
}
