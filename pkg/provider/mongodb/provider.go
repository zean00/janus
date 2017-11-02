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
}

// Provide allows the file provider to provide configurations to janus
// using the given configuration channel.
func (p *Provider) Provide(configChan chan<- types.ConfigMessage) error {
	session, err := p.initMongoSession()
	if err != nil {
		return err
	}

	p.Session = session
	repo := NewMongoRepository(session)
	backends, err := repo.FindAll()
	if err != nil {
		return err
	}

	sendConfigToChannel(configChan, backends)
	return nil
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

func sendConfigToChannel(configChan chan<- types.ConfigMessage, backends []*types.Backend) {
	configChan <- types.ConfigMessage{
		ProviderName:  "mongodb",
		Configuration: backends,
	}
}
