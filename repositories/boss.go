package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
)

type boss struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type BossRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (boss *models.Monster, err error)
	QueryDocuments(args *[]models.QueryArg) (*[]models.Monster, error)
}

func NewBossRepository(log loggo.Logger, ds datasource.Datasource) BossRepository {
	return &boss{
		log: log,
		ds:  ds,
	}
}

func (b *boss) InsertDocument(id *string, data interface{}) (*string, error) {
	if id != nil {
		_, err := b.ds.InsertDocumentWithID(globals.BOSSES, *id, data)
		if err != nil {
			b.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := b.ds.InsertDocument(globals.BOSSES, data)
		if err != nil {
			b.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (b *boss) ReadDocument(id string) (boss *models.Monster, err error) {
	doc, err := b.ds.ReadDocument(globals.BOSSES, id)
	if err != nil {
		b.log.Errorf("error reading classes: %v", err)
		return nil, err
	}
	boss = &models.Monster{}
	err = doc.DataTo(&boss)
	if err != nil {
		b.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return boss, nil
}

func (b *boss) QueryDocuments(args *[]models.QueryArg) (*[]models.Monster, error) {
	docs, err := b.ds.QueryCollection(globals.BOSSES, args)
	if err != nil {
		b.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	var bosses []models.Monster
	for _, doc := range docs {
		boss := models.Monster{}
		err := doc.DataTo(&boss)
		if err != nil {
			b.log.Errorf("error converting doc to area with error: %v", err)
			return nil, err
		}
		bosses = append(bosses, boss)
	}

	return &bosses, nil
}
