package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"lataleBotService/utils"
	"time"
)

type level struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type LevelRepository interface {
	InsertDocument(id *string, level models.Level) (*models.Level, error)
	ReadDocument(id string) (level *models.Level, err error)
	QueryDocuments(collection string, args *[]models.QueryArg) (map[string]*models.Level, error)
	UpdateDocument(docId string, data interface{}) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewLevelRepo(log loggo.Logger, ds datasource.Datasource) LevelRepository {
	return &level{
		log: log,
		ds:  ds,
	}
}

func (l *level) InsertDocument(id *string, level models.Level) (*models.Level, error) {
	if id != nil {
		//err := l.ds.OpenConnection()
		//if err != nil {
		//	l.log.Errorf("error opening connection to the datasource: %v", err)
		//	return nil, err
		//}
		//defer l.ds.CloseConnection()
		//_, err = l.ds.InsertDocumentWithID(globals.LEVELS, *id, level)
		//if err != nil {
		//	l.log.Errorf("error Inserting Document: %v", err)
		//	return nil, err
		//}
		return &level, nil
	} else {
		_, err := l.ds.InsertDocument(globals.LEVELS, level)
		if err != nil {
			l.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return &level, nil
	}
}

func (l *level) ReadDocument(id string) (level *models.Level, err error) {
	//err = l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("error opening ds connection: %v", err)
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()

	doc, err := l.ds.ReadDocument(globals.LEVELS, id)
	if err != nil {
		l.log.Errorf("error reading level: %v", err)
		return nil, err
	}
	l.log.Debugf("doc: %v", doc)
	level = &models.Level{}
	err = doc.DataTo(level)
	if err != nil {
		l.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return level, nil
}

func (l *level) QueryDocuments(collection string, args *[]models.QueryArg) (map[string]*models.Level, error) {
	//err := l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("failed to open datasource connection")
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()
	docs, err := l.ds.QueryCollection(collection, args)
	if err != nil {
		l.log.Errorf("error querying for documents with error: %v", err)
		return nil, err
	}
	levels := make(map[string]*models.Level)
	for _, doc := range docs {
		if doc.Ref.ID != globals.LEVELCAP {
			level := models.Level{}
			err := doc.DataTo(&level)
			if err != nil {
				l.log.Errorf("error converting doc to profile with error: %v", err)
				return nil, err
			}
			levels[utils.String(level.Value)] = &level
		}
	}
	return levels, nil
}

func (l *level) UpdateDocument(docId string, data interface{}) (*time.Time, error) {
	//err := l.ds.OpenConnection()
	//if err != nil {
	//	l.log.Errorf("failed to open datasource connection")
	//	return nil, err
	//}
	//defer l.ds.CloseConnection()
	updateTS, err := l.ds.UpdateDocument(globals.LEVELS, docId, data)
	if err != nil {
		l.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, data, err)
		return nil, err
	}
	return updateTS, nil
}

func (l *level) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
