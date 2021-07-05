package models

type SetBonus struct {
	Name              string        `json:"name" firestore:"name"`
	Id                string        `json:"id" firestore:"id"`
	CurrentlyEquipped int32         `json:"currentlyEquipped" firestore:"currentlyEquipped"`
	RequiredPieces    int32         `json:"requiredPieces" firestore:"requiredPieces"`
	Bonus             *StatModifier `json:"setBonuses" firestore:"setBonuses"`
}
