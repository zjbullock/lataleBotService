package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type character struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type ClassRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (class *models.JobClass, err error)
	QueryDocuments(args []models.QueryArg) error
	UpdateDocument(docId string) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewClassRepo(log loggo.Logger, ds datasource.Datasource) ClassRepository {
	return &character{
		log: log,
		ds:  ds,
	}
}

func (c *character) InsertDocument(id *string, data interface{}) (*string, error) {
	if id != nil {
		err := c.ds.OpenConnection()
		if err != nil {
			c.log.Errorf("error opening connection to the datasource: %v", err)
			return nil, err
		}
		defer c.ds.CloseConnection()
		_, err = c.ds.InsertDocumentWithID(globals.CLASSES, *id, data)
		if err != nil {
			c.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := c.ds.InsertDocument(globals.CLASSES, data)
		if err != nil {
			c.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (c *character) ReadDocument(id string) (classInfo *models.JobClass, err error) {
	err = c.ds.OpenConnection()
	if err != nil {
		c.log.Errorf("error opening ds connection: %v", err)
		return nil, err
	}
	defer c.ds.CloseConnection()

	doc, err := c.ds.ReadDocument(globals.CLASSES, id)
	if err != nil {
		c.log.Errorf("error reading user: %v", err)
		return nil, err
	}
	var class models.JobClass
	err = doc.DataTo(&class)
	if err != nil {
		c.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return &class, nil
}

func (c *character) QueryDocuments(args []models.QueryArg) error {
	panic("implement me")
}

func (c *character) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (c *character) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
