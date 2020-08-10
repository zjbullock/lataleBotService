package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/models"
	"time"
)

type user struct {
	log loggo.Logger
	ds  datasource.Datasource
}

func NewUserRepo(log loggo.Logger, ds datasource.Datasource) Repository {
	return &user{
		log: log,
		ds:  ds,
	}
}

func (*user) InsertDocument(data interface{}) (*string, error) {
	panic("implement me")
}

func (*user) ReadDocument(id string) (error) {
	panic("implement me")
}

func (*user) QueryDocuments(args []models.QueryArg) (error) {
	panic("implement me")
}

func (*user) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (*user) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}