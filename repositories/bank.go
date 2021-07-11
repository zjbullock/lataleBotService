package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type bank struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type BankRepository interface {
	InsertDocument(userId string, data interface{}) (*string, error)
	ReadDocument(id string) (item *models.Inventory, err error)
	QueryDocuments(args *[]models.QueryArg) (*models.Inventory, error)
	QueryForDocument(args *[]models.QueryArg) (*models.Inventory, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewBankRepo(log loggo.Logger, ds datasource.Datasource) BankRepository {
	return &bank{
		log: log,
		ds:  ds,
	}
}

func (b *bank) InsertDocument(userId string, data interface{}) (*string, error) {
	_, err := b.ds.InsertDocumentWithID(globals.BANK, userId, data)
	if err != nil {
		b.log.Errorf("error Inserting Document: %v", err)
		return nil, err
	}
	return &userId, nil
}

func (b *bank) ReadDocument(id string) (item *models.Inventory, err error) {
	doc, err := b.ds.ReadDocument(globals.BANK, id)
	if err != nil {
		b.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	var bank models.Inventory
	err = doc.DataTo(&bank)
	if err != nil {
		b.log.Errorf("error converting doc to item with error: %v", err)
		return nil, err
	}

	return &bank, nil
}

func (b *bank) QueryForDocument(args *[]models.QueryArg) (*models.Inventory, error) {
	panic("implement me")
}

func (b *bank) QueryDocuments(args *[]models.QueryArg) (*models.Inventory, error) {
	panic("implement me")
}

func (b *bank) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	updateTime, err := b.ds.UpdateDocument(globals.BANK, docId, data)
	if err != nil {
		b.log.Errorf("error updating bank")
		return nil, err
	}
	return updateTime, nil
}

func (b *bank) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
