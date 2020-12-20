package models

type Level struct {
	Value int32 `json:"value" firestore:"value"`
	Exp   int64 `json:"exp" firestore:"exp"`
}
