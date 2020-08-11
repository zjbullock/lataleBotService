package models

type ClassInfo struct {
	Name      string    `json:"name" firestore:"name"`
	Level     int32     `json:"level" firestore:"level"`
	Exp       int32     `json:"exp" firestore:"exp"`
	Equipment Equipment `json:"equipment" firestore:"equipment"`
}
