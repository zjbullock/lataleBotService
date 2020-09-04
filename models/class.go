package models

type JobClass struct {
	Name             string       `json:"name" firestore:"name"`
	LevelRequirement int64        `json:"levelRequirement" firestore:"levelRequirement"`
	ClassRequirement *string      `json:"classRequirement" firestore:"classRequirement"`
	Tier             int32        `json:"tier" firestore:"tier"`
	Weapons          []Weapon     `json:"weapons" firestore:"weapons"`
	Stats            StatModifier `json:"stats" firestore:"stats"`
	Description      string       `json:"description" firestore:"description"`
}
