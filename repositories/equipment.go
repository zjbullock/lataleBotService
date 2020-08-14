package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type equipment struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type EquipmentRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (equipment *models.EquipmentSheet, err error)
	QueryDocuments(args []models.QueryArg) error
	UpdateDocument(docId string) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewEquipmentRepo(log loggo.Logger, ds datasource.Datasource) EquipmentRepository {
	return &equipment{
		log: log,
		ds:  ds,
	}
}

func (e *equipment) InsertDocument(id *string, data interface{}) (*string, error) {
	if id != nil {
		err := e.ds.OpenConnection()
		if err != nil {
			e.log.Errorf("error opening connection to the datasource: %v", err)
			return nil, err
		}
		defer e.ds.CloseConnection()
		_, err = e.ds.InsertDocumentWithID(globals.EQUIPMENT, *id, data)
		if err != nil {
			e.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := e.ds.InsertDocument(globals.EQUIPMENT, data)
		if err != nil {
			e.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (e *equipment) ReadDocument(id string) (equipment *models.EquipmentSheet, err error) {
	err = e.ds.OpenConnection()
	if err != nil {
		e.log.Errorf("error opening ds connection: %v", err)
		return nil, err
	}
	defer e.ds.CloseConnection()

	doc, err := e.ds.ReadDocument(globals.EQUIPMENT, id)
	if err != nil {
		e.log.Errorf("error reading equipment: %v with id: %s", err, id)
		return nil, err
	}
	var equips models.EquipmentSheet
	err = doc.DataTo(&equips)
	if err != nil {
		e.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	equips.ID = doc.Ref.ID
	return &equips, nil
}

func (e *equipment) QueryDocuments(args []models.QueryArg) error {
	panic("implement me")
}

func (e *equipment) UpdateDocument(docId string) (*time.Time, error) {
	panic("implement me")
}

func (e *equipment) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
