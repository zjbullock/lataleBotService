package models

type JobClass struct {
	Name             string       `json:"name" firestore:"name"`
	LevelRequirement int32        `json:"levelRequirement" firestore:"levelRequirement"`
	ClassRequirement *string      `json:"classRequirement" firestore:"classRequirement"`
	Weapons          []Weapon     `json:"weapons" firestore:"weapons"`
	Stats            StatModifier `json:"stats" firestore:"stats"`
	Description      string       `json:"description" firestore:"description"`
}
