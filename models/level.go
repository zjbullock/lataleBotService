package models

type Level struct {
	Value int32 `json:"value" firestore:"value"`
	Exp   int32 `json:"exp" firestore:"exp"`
}
