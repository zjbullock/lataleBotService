package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type area struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type AreasRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (area *models.Area, err error)
	QueryDocuments(args *[]models.QueryArg) (*[]models.Area, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewAreaRepo(log loggo.Logger, ds datasource.Datasource) AreasRepository {
	return &area{
		log: log,
		ds:  ds,
	}
}

func (a *area) InsertDocument(id *string, data interface{}) (*string, error) {
	if id != nil {
		_, err := a.ds.InsertDocumentWithID(globals.AREAS, *id, data)
		if err != nil {
			a.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := a.ds.InsertDocument(globals.AREAS, data)
		if err != nil {
			a.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (a *area) ReadDocument(id string) (area *models.Area, err error) {
	doc, err := a.ds.ReadDocument(globals.AREAS, id)
	if err != nil {
		a.log.Errorf("error reading user: %v", err)
		return nil, err
	}
	a.log.Debugf("doc: %v", doc)
	area = &models.Area{}
	err = doc.DataTo(area)
	if err != nil {
		a.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return area, nil
}

func (a *area) QueryDocuments(args *[]models.QueryArg) (*[]models.Area, error) {
	docs, err := a.ds.QueryCollection(globals.AREAS, args)
	if err != nil {
		a.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	var areas []models.Area
	for _, doc := range docs {
		area := models.Area{}
		err := doc.DataTo(&area)
		if err != nil {
			a.log.Errorf("error converting doc to area with error: %v", err)
			return nil, err
		}
		areas = append(areas, area)
	}

	return &areas, nil
}

func (a *area) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	updateTS, err := a.ds.UpdateDocument(globals.AREAS, docId, data)
	if err != nil {
		a.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, data, err)
		return nil, err
	}
	return updateTS, nil
}

func (a *area) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
