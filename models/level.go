package models

type Level struct {
	Value int32   `json:"value" firestore:"value"`
	Exp   float64 `json:"exp" firestore:"exp"`
}
