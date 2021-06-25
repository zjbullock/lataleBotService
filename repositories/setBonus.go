package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type setBonus struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type SetBonusRepository interface {
	InsertDocument(id string, data interface{}) (*string, error)
	ReadDocument(id string) (setBonus *models.SetBonus, err error)
	QueryDocuments(args *[]models.QueryArg) ([]models.SetBonus, error)
	QueryForDocument(args *[]models.QueryArg) (*models.SetBonus, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewSetBonusRepo(log loggo.Logger, ds datasource.Datasource) SetBonusRepository {
	return &setBonus{
		log: log,
		ds:  ds,
	}
}

func (s *setBonus) InsertDocument(id string, data interface{}) (*string, error) {
	_, err := s.ds.InsertDocumentWithID(globals.SETBONUS, id, data)
	if err != nil {
		s.log.Errorf("error Inserting Document: %v", err)
		return nil, err
	}
	return &id, nil
}

func (s *setBonus) ReadDocument(id string) (setBonus *models.SetBonus, err error) {
	doc, err := s.ds.ReadDocument(globals.SETBONUS, id)
	if err != nil {
		s.log.Errorf("failed to get set bonus from collection: %v", err)
		return nil, err
	}
	setBonus = &models.SetBonus{}
	err = doc.DataTo(&setBonus)
	if err != nil {
		s.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return setBonus, nil
}

func (s *setBonus) QueryForDocument(args *[]models.QueryArg) (*models.SetBonus, error) {
	panic("implement me")
}

func (s *setBonus) QueryDocuments(args *[]models.QueryArg) ([]models.SetBonus, error) {
	panic("implement me")
}

func (s *setBonus) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	panic("implement me")
}

func (s *setBonus) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
