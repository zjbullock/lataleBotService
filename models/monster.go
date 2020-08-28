package models

type Monster struct {
	Name      string       `json:"name" firestore:"name"`
	Level     int32        `json:"level" firestore:"level"`
	Ely       int32        `json:"ely" firestore:"ely"`
	Exp       int32        `json:"exp" firestore:"exp"`
	Rank      int32        `json:"rank" firestore:"rank"`
	Stats     StatModifier `json:"stats" firestore:"stats"`
	BossBonus *int         `json:"isBoss" firestore:"isBoss"`
	Skills    *[]string    `json:"skills,omitempty" firestore:"skills,omitempty"`
}
