package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/models"
	"time"
)

type character struct {
	log loggo.Logger
	ds  datasource.Datasource
}

func NewCharacterRepo(log loggo.Logger, ds datasource.Datasource) Repository {
	return &character{
		log: log,
		ds:  ds,
	}
}

func (*character) InsertDocument(data interface{}) (*string, error) {
	panic("implement me")
}

func (*character) ReadDocument(id string) (error) {
	panic("implement me")
}

func (*character) QueryDocuments(args []models.QueryArg) (error) {
	panic("implement me")
}

func (*character) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (*character) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
