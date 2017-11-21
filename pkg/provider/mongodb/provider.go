package mongodb

import (
	"github.com/containous/traefik/log"
	"github.com/hellofresh/janus/pkg/types"
	mgo "gopkg.in/mgo.v2"
)

// Provider holds configurations of the provider.
type Provider struct {
	DSN     string `envconfig:"DATABASE_DSN"`
	Session *mgo.Session
	repo    *MongoRepository
}

// Provide allows the file provider to provide configurations to janus
// using the given configuration channel.
func (p *Provider) Provide(configChan chan<- types.ConfigMessage, configChangeChan chan types.ConfigurationEvent) error {
	session, err := p.initMongoSession()
	if err != nil {
		return err
	}

	p.Session = session
	p.repo = NewMongoRepository(session)
	backends, err := p.repo.FindAll()
	if err != nil {
		return err
	}

	go p.listenForChanges(configChangeChan)
	sendConfigToChannel(configChan, &types.Configuration{Backends: backends})

	return nil
}

func (p *Provider) listenForChanges(configChangeChan <-chan types.ConfigurationEvent) {
	for {
		changeConfig := <-configChangeChan
		switch changeConfig.Type {
		case types.ConfigurationChanged:
			err := p.repo.Add(changeConfig.Backend)
			if err != nil {
				log.WithError(err).Error("Could not add the backend configuration")
			}
		case types.ConfigurationRemoved:
			err := p.repo.Remove(changeConfig.Backend.Name)
			if err != nil {
				log.WithError(err).Error("Could not remove the backend configuration")
			}
		}
	}
}

func (p *Provider) initMongoSession() (*mgo.Session, error) {
	log.Debug("MongoDB configuration chosen")

	log.WithField("dsn", p.DSN).Debug("Trying to connect to MongoDB...")
	session, err := mgo.Dial(p.DSN)
	if err != nil {
		return nil, err
	}

	log.Debug("Connected to MongoDB")
	session.SetMode(mgo.Monotonic, true)
	return session, nil
}

func sendConfigToChannel(configChan chan<- types.ConfigMessage, configurations *types.Configuration) {
	configChan <- types.ConfigMessage{
		ProviderName:  "mongodb",
		Configuration: configurations,
	}
}
