package models

type NewUserResponse struct {
	ID      *string `json:"id" firestore:"id"`
	Message *string `json:"message" firestore:"message"`
}
