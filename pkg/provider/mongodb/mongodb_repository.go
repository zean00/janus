package mongodb

import (
	"errors"

	"github.com/hellofresh/janus/pkg/types"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrAPIDefinitionNotFound is used when the api was not found in the datastore
	ErrAPIDefinitionNotFound = errors.New("api definition not found")

	// ErrAPINameExists is used when the API name is already registered on the datastore
	ErrAPINameExists = errors.New("api name is already registered")

	// ErrAPIListenPathExists is used when the API listen path is already registered on the datastore
	ErrAPIListenPathExists = errors.New("api listen path is already registered")
)

const (
	collectionName string = "api_specs"
)

// MongoRepository represents a mongodb repository
type MongoRepository struct {
	session *mgo.Session
}

// NewMongoRepository creates a mongo API definition repo
func NewMongoRepository(session *mgo.Session) *MongoRepository {
	return &MongoRepository{session}
}

// FindAll fetches all the API definitions available
func (r *MongoRepository) FindAll() ([]*types.Backend, error) {
	result := []*types.Backend{}
	session, coll := r.getSession()
	defer session.Close()

	err := coll.Find(nil).All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindByName find an API definition by name
func (r *MongoRepository) FindByName(name string) (*types.Backend, error) {
	return r.findOneByQuery(bson.M{"name": name})
}

// FindByListenPath find an API definition by proxy listen path
func (r *MongoRepository) FindByListenPath(path string) (*types.Backend, error) {
	return r.findOneByQuery(bson.M{"proxy.listen_path": path})
}

func (r *MongoRepository) findOneByQuery(query interface{}) (*types.Backend, error) {
	var result = types.NewBackend()
	session, coll := r.getSession()
	defer session.Close()

	err := coll.Find(query).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrAPIDefinitionNotFound
		}
		return nil, err
	}

	return result, err
}

// Exists searches an existing API definition by its listen_path
func (r *MongoRepository) Exists(def *types.Backend) (bool, error) {
	_, err := r.FindByName(def.Name)
	if nil != err && err != ErrAPIDefinitionNotFound {
		return false, err
	} else if err != ErrAPIDefinitionNotFound {
		return true, ErrAPINameExists
	}

	_, err = r.FindByListenPath(def.Proxy.ListenPath)
	if nil != err && err != ErrAPIDefinitionNotFound {
		return false, err
	} else if err != ErrAPIDefinitionNotFound {
		return true, ErrAPIListenPathExists
	}

	return false, nil
}

// Add adds an API definition to the repository
func (r *MongoRepository) Add(definition *types.Backend) error {
	session, coll := r.getSession()
	defer session.Close()

	isValid, err := definition.Validate()
	if false == isValid && err != nil {
		log.WithError(err).Error("Validation errors")
		return err
	}

	_, err = coll.Upsert(bson.M{"name": definition.Name}, definition)
	if err != nil {
		log.WithField("name", definition.Name).Error("There was an error adding the resource")
		return err
	}

	log.WithField("name", definition.Name).Debug("Resource added")
	return nil
}

// Remove removes an API definition from the repository
func (r *MongoRepository) Remove(name string) error {
	session, coll := r.getSession()
	defer session.Close()

	err := coll.Remove(bson.M{"name": name})
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrAPIDefinitionNotFound
		}
		log.WithField("name", name).Error("There was an error removing the resource")
		return err
	}

	log.WithField("name", name).Debug("Resource removed")
	return nil
}

// FindValidAPIHealthChecks retrieves all apis that has health check configured
func (r *MongoRepository) FindValidAPIHealthChecks() ([]*types.Backend, error) {
	result := []*types.Backend{}
	session, coll := r.getSession()
	defer session.Close()

	err := coll.Find(bson.M{"health_check.url": bson.M{"$exists": true}}).All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MongoRepository) getSession() (*mgo.Session, *mgo.Collection) {
	session := r.session.Copy()
	coll := session.DB("").C(collectionName)

	return session, coll
}
