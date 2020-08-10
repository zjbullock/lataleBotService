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

func NewAreaRepo(log loggo.Logger, ds datasource.Datasource) Repository {
	return &area{
		log: log,
		ds:  ds,
	}
}

func (*area) InsertDocument(data interface{}) (*string, error) {
	panic("implement me")
}

func (*area) ReadDocument(id string) (error) {
	panic("implement me")
}

func (*area) QueryDocuments(args []models.QueryArg) (error) {
	panic("implement me")
}

func (*area) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (*area) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
