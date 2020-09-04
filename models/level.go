package models

type Level struct {
	Value int64 `json:"value" firestore:"value"`
	Exp   int64 `json:"exp" firestore:"exp"`
}
