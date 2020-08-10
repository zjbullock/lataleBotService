package repositories

import (
	"lataleBotService/models"
	"time"
)

type Repository interface {
	InsertDocument(data interface{}) (*string, error)
	ReadDocument(id string) (error)
	QueryDocuments(args []models.QueryArg) (error)
	UpdateDocument(docId string) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}