package datasource

import (
	. "cloud.google.com/go/firestore"
	"context"
	"github.com/juju/loggo"
	"google.golang.org/api/iterator"
	"lataleBotService/models"
	"time"
)

type Datasource interface {
	//OpenConnection() error
	//CloseConnection() error
	InsertDocument(collection string, data interface{}) (*string, error)
	DeleteDocument(collection, id string) (*time.Time, error)
	InsertDocumentWithID(collection, id string, data interface{}) (*time.Time, error)
	UpdateDocument(collection, profileId string, data interface{}) (*time.Time, error)
	UpdateDocumentField(collection, profileId string, updates []Update) (*time.Time, error)
	QueryCollection(collection string, args *[]models.QueryArg) ([]*DocumentSnapshot, error)
	ReadDocument(collection, id string) (*DocumentSnapshot, error)
}

type fireStoreDB struct {
	log    loggo.Logger
	ctx    context.Context
	Client *Client
}

func NewDataSource(l loggo.Logger, ctx context.Context, client *Client) Datasource {
	return &fireStoreDB{
		log:    l,
		ctx:    ctx,
		Client: client,
	}
}

func (f *fireStoreDB) DeleteDocument(collection, id string) (*time.Time, error) {
	wr, err := f.Client.Collection(collection).Doc(id).Delete(f.ctx)
	if err != nil {
		f.log.Errorf("error deleting document %s in collection: %s with error: %v", id, collection, err)
		return nil, err
	}
	return &wr.UpdateTime, nil
}

func (f *fireStoreDB) InsertDocumentWithID(collection, id string, data interface{}) (*time.Time, error) {
	f.log.Debugf("collection: %s, id: %s, data: %v", collection, id, data)
	wr, err := f.Client.Doc(collection+"/"+id).Set(f.ctx, data)
	if err != nil {
		f.log.Errorf("error inserting document into collection: %s with error: %v", collection, err)
		return nil, err
	}
	f.log.Debugf("wr: %v", wr)
	return &wr.UpdateTime, nil
}

func (f *fireStoreDB) InsertDocument(collection string, data interface{}) (*string, error) {
	doc, _, err := f.Client.Collection(collection).Add(f.ctx, data)
	if err != nil {
		f.log.Errorf("error inserting document into collection: %s with error: %v", collection, err)
		return nil, err
	}
	return &doc.ID, nil
}

func (f *fireStoreDB) UpdateDocument(collection, profileId string, data interface{}) (*time.Time, error) {
	f.log.Debugf("collection: %s, id: %s, data: %v", collection, profileId, data)

	res, err := f.Client.Collection(collection).Doc(profileId).Set(f.ctx, data)
	if err != nil {
		f.log.Errorf("error setting document: %s in collection :%s with error: %v", profileId, collection, err)
		return nil, err
	}
	return &res.UpdateTime, nil
}

func (f *fireStoreDB) UpdateDocumentField(collection, profileId string, updates []Update) (*time.Time, error) {
	res, err := f.Client.Collection(collection).Doc(profileId).Update(f.ctx, updates)
	if err != nil {
		f.log.Errorf("error updating document %s in collection: %s with error: %v", profileId, collection, err)
		return nil, err
	}
	return &res.UpdateTime, nil
}

func (f *fireStoreDB) QueryCollection(collection string, args *[]models.QueryArg) ([]*DocumentSnapshot, error) {
	var iter *DocumentIterator
	if args != nil {
		q := f.Client.Collection(collection).Query
		for _, arg := range *args {
			q = q.Where(arg.Path, arg.Op, arg.Value)
		}
		q.Documents(f.ctx)
		iter = q.Documents(f.ctx)
	} else {
		q := f.Client.Collection(collection)
		q.Documents(f.ctx)
		iter = q.Documents(f.ctx)
	}

	defer iter.Stop()
	var docs []*DocumentSnapshot
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			f.log.Errorf("error iterating through queried documents with error: %v", err)
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (f *fireStoreDB) ReadDocument(collection, id string) (*DocumentSnapshot, error) {
	doc, err := f.Client.Collection(collection).Doc(id).Get(f.ctx)
	if err != nil {
		f.log.Errorf("error reading document with id: %s with error: %v", id, err)
		return nil, err
	}
	return doc, nil
}
