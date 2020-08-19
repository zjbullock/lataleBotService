package repositories

import (
	"github.com/juju/loggo"
	"lataleBotService/datasource"
	"lataleBotService/globals"
	"lataleBotService/models"
	"time"
)

type user struct {
	log loggo.Logger
	ds  datasource.Datasource
}

type UserRepository interface {
	InsertDocument(id *string, data interface{}) (*string, error)
	ReadDocument(id string) (user *models.User, err error)
	QueryDocuments(args []models.QueryArg) error
	UpdateDocument(docId string, user *models.User) (*time.Time, error)
	UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error)
}

func NewUserRepo(log loggo.Logger, ds datasource.Datasource) UserRepository {
	return &user{
		log: log,
		ds:  ds,
	}
}

func (u *user) InsertDocument(id *string, data interface{}) (*string, error) {
	if id != nil {
		err := u.ds.OpenConnection()
		if err != nil {
			u.log.Errorf("error opening connection to the datasource: %v", err)
			return nil, err
		}
		defer u.ds.CloseConnection()
		_, err = u.ds.InsertDocumentWithID(globals.USERS, *id, data)
		if err != nil {
			u.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	} else {
		id, err := u.ds.InsertDocument(globals.USERS, data)
		if err != nil {
			u.log.Errorf("error Inserting Document: %v", err)
			return nil, err
		}
		return id, nil
	}
}

func (u *user) ReadDocument(id string) (userInfo *models.User, err error) {
	err = u.ds.OpenConnection()
	if err != nil {
		u.log.Errorf("error opening ds connection: %v", err)
		return nil, err
	}
	defer u.ds.CloseConnection()

	doc, err := u.ds.ReadDocument(globals.USERS, id)
	if err != nil {
		u.log.Errorf("error reading user: %v", err)
		return nil, err
	}
	u.log.Debugf("doc: %v", doc)
	var user models.User
	err = doc.DataTo(&user)
	if err != nil {
		u.log.Errorf("error converting doc: %v", err)
		return nil, err
	}
	return &user, nil
}

func (u *user) QueryDocuments(args []models.QueryArg) error {
	panic("implement me")
}

func (u *user) UpdateDocument(docId string, user *models.User) (*time.Time, error) {
	err := u.ds.OpenConnection()
	if err != nil {
		u.log.Errorf("failed to open datasource connection")
		return nil, err
	}
	defer u.ds.CloseConnection()
	updateTS, err := u.ds.UpdateDocument(globals.USERS, docId, user)
	if err != nil {
		u.log.Errorf("failed to update doc: %s with data: %v.  Received err: %v", docId, user, err)
		return nil, err
	}
	return updateTS, nil
}

func (u *user) UpdateDocumentFields(docId, field string, data interface{}) (*time.Time, error) {
	panic("implement me")
}
