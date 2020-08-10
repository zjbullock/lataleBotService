package models

type JobClass struct {
	Name             string   `json:"name" firestore:"name"`
	LevelRequirement int      `json:"levelRequirement" firestore:"levelRequirement"`
	Weapons          []Weapon `json:"weapons" firestore:"weapons"`
	Stats            Stat     `json:"stats" firestore:"stats"`
}
