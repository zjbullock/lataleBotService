package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/models"
	"time"
)

type area struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type AreasRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (data interface{}, err error)
	QueryDocuments(args []models.QueryArg) error
	UpdateDocument(docId string) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewAreaRepo(log loggo.Logger, ds datasource.Datasource) AreasRepository {
	return &area{
		log: log,
		ds:  ds,
	}
}

func (*area) InsertDocument(id *string, data interface{}) (*string, error) {
	panic("implement me")
}

func (*area) ReadDocument(id string) (data interface{}, err error) {
	panic("implement me")
}

func (*area) QueryDocuments(args []models.QueryArg) error {
	panic("implement me")
}

func (*area) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (*area) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
