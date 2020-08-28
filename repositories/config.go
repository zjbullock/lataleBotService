package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type config struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type ConfigRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (config map[string]*int, err error)
	QueryDocuments(args *[]models.QueryArg) error
	UpdateDocument(data interface{}, docId string) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewConfigRepo(log loggo.Logger, ds datasource.Datasource) ConfigRepository {
	return &config{
		log: log,
		ds:  ds,
	}
}

func (c *config) InsertDocument(id *string, data interface{}) (*string, error) {
	panic("implement me")
}

func (c *config) ReadDocument(id string) (config map[string]*int, err error) {
	doc, err := c.ds.ReadDocument(globals.CONFIG, id)
	if err != nil {
		c.log.Errorf("error reading equipment: %v with id: %s", err, id)
		return nil, err
	}
	var expConfig map[string]*int
	err = doc.DataTo(&expConfig)
	if err != nil {
		c.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return expConfig, nil
}

func (c *config) QueryDocuments(args *[]models.QueryArg) error {
	panic("implement me")
}

func (c *config) UpdateDocument(data interface{}, docId string) (*time.Time, error) {
	updateTS, err := c.ds.UpdateDocument(globals.CONFIG, docId, &data)
	if err != nil {
		c.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, data, err)
		return nil, err
	}
	return updateTS, nil
}

func (c *config) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
