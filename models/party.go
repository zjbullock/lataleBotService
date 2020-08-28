package models

type Party struct {
	ID      *string  `json:"id" firestore:"id"`
	Leader  string   `json:"leader" firestore:"leader"`
	Members []string `json:"members" firestore:"members"`
}
