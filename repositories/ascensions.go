package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/utils"
	"time"
)

type ascension struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type AscensionRepository interface {
	InsertDocument(id *string, level models.Level) (*models.Level, error)
	ReadDocument(id string) (level *models.Level, err error)
	QueryDocuments(collection string, args *[]models.QueryArg) (map[string]*models.Level, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewAscensionRepository(log loggo.Logger, ds datasource.Datasource) AscensionRepository {
	return &ascension{
		log: log,
		ds:  ds,
	}
}

func (a *ascension) InsertDocument(id *string, level models.Level) (*models.Level, error) {
	if id != nil {
		_, err := a.ds.InsertDocumentWithID(globals.ASCENSION, *id, level)
		if err != nil {
			a.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return &level, nil
	} else {
		_, err := a.ds.InsertDocument(globals.ASCENSION, level)
		if err != nil {
			a.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return &level, nil
	}
}

func (a *ascension) ReadDocument(id string) (level *models.Level, err error) {
	//err = l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("error opening ds connection: %v", err)
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()

	doc, err := a.ds.ReadDocument(globals.ASCENSION, id)
	if err != nil {
		a.log.Errorf("error reading level: %v", err)
		return nil, err
	}
	level = &models.Level{}
	err = doc.DataTo(level)
	if err != nil {
		a.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return level, nil
}

func (a *ascension) QueryDocuments(collection string, args *[]models.QueryArg) (map[string]*models.Level, error) {
	//err := l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("failed to open datasource connection")
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()
	docs, err := a.ds.QueryCollection(collection, args)
	if err != nil {
		a.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	levels := make(map[string]*models.Level)
	for _, doc := range docs {
		if doc.Ref.ID != globals.LEVELCAP {
			level := models.Level{}
			err := doc.DataTo(&level)
			if err != nil {
				a.log.Errorf("error converting doc to profile with error: %v", err)
				return nil, err
			}
			levels[utils.ThirtyTwoBitIntToString(level.Value)] = &level
		}
	}
	return levels, nil
}

func (a *ascension) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	//err := l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("failed to open datasource connection")
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()
	updateTS, err := a.ds.UpdateDocument(globals.ASCENSION, docId, data)
	if err != nil {
		a.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, data, err)
		return nil, err
	}
	return updateTS, nil
}

func (a *ascension) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
