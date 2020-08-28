package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type party struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type PartyRepository interface {
	InsertDocument(id *string, party *models.Party) (*string, error)
	DeleteDocument(id string) (err error)
	ReadDocument(id string) (party *models.Party, err error)
	QueryDocuments(args *[]models.QueryArg) (*models.Party, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId string, data interface{}) (*time.Time, error)
}

func NewPartiesRepo(log loggo.Logger, ds datasource.Datasource) PartyRepository {
	return &party{
		log: log,
		ds:  ds,
	}
}

func (p *party) InsertDocument(id *string, party *models.Party) (*string, error) {
	if id != nil {
		_, err := p.ds.InsertDocumentWithID(globals.PARTY, *id, party)
		if err != nil {
			p.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := p.ds.InsertDocument(globals.PARTY, party)
		if err != nil {
			p.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		party.ID = id
		_, err = p.ds.UpdateDocument(globals.PARTY, *id, party)
		if err != nil {
			p.log.Errorf("error Updating Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (p *party) DeleteDocument(id string) (err error) {
	_, err = p.ds.DeleteDocument(globals.PARTY, id)
	if err != nil {
		p.log.Errorf("error deleting document with the specified id: %v", err)
		return err
	}
	return nil
}

func (p *party) ReadDocument(id string) (party *models.Party, err error) {
	doc, err := p.ds.ReadDocument(globals.PARTY, id)
	if err != nil {
		p.log.Errorf("error reading user: %v", err)
		return nil, err
	}
	p.log.Debugf("doc: %v", doc)
	party = &models.Party{}
	err = doc.DataTo(party)
	if err != nil {
		p.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return party, nil
}

func (p *party) QueryDocuments(args *[]models.QueryArg) (*models.Party, error) {
	docs, err := p.ds.QueryCollection(globals.PARTY, args)
	if err != nil {
		p.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	party := &models.Party{}
	for _, doc := range docs {
		err := doc.DataTo(party)
		if err != nil {
			p.log.Errorf("error converting doc to party with error: %v", err)
			return nil, err
		}
	}
	p.log.Debugf("party: %v", party)
	return party, nil
}

func (p *party) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	updateTS, err := p.ds.UpdateDocument(globals.PARTY, docId, data)
	if err != nil {
		p.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, data, err)
		return nil, err
	}
	return updateTS, nil
}

func (p *party) UpdateDocumentFields(docId string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
