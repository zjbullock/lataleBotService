package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type item struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type ItemRepository interface {
	InsertDocument(data interface{}) (*string, error)
	ReadDocument(id string) (item *models.Item, err error)
	QueryDocuments(args *[]models.QueryArg) ([]models.Item, error)
	QueryForDocument(args *[]models.QueryArg) (*models.Item, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewItemRepo(log loggo.Logger, ds datasource.Datasource) ItemRepository {
	return &item{
		log: log,
		ds:  ds,
	}
}

func (i *item) InsertDocument(data interface{}) (*string, error) {
	id, err := i.ds.InsertDocument(globals.ITEM, data)
	if err != nil {
		i.log.Errorf("error Inserting Document: %v", err)
		return nil, err
	}
	return id, nil
}

func (i *item) ReadDocument(id string) (item *models.Item, err error) {
	panic("implement me")
}

func (i *item) QueryForDocument(args *[]models.QueryArg) (*models.Item, error) {
	docs, err := i.ds.QueryCollection(globals.ITEM, args)
	if err != nil {
		i.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	var items []models.Item
	for _, doc := range docs {
		item := models.Item{}
		err := doc.DataTo(&item)
		if err != nil {
			i.log.Errorf("error converting doc to area with error: %v", err)
			return nil, err
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, nil
	}

	return &items[0], nil
}

func (i *item) QueryDocuments(args *[]models.QueryArg) ([]models.Item, error) {
	docs, err := i.ds.QueryCollection(globals.ITEM, args)
	if err != nil {
		i.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	var items []models.Item
	for _, doc := range docs {
		item := models.Item{}
		err := doc.DataTo(&item)
		if err != nil {
			i.log.Errorf("error converting doc to area with error: %v", err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (i *item) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	panic("implement me")
}

func (i *item) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
