package main

import (
	"net/url"

	"github.com/hellofresh/janus/pkg/config"
	tracerfactory "github.com/hellofresh/janus/pkg/opentracing"
	"github.com/hellofresh/janus/pkg/store"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var (
	globalConfig *config.Specification
	storage      store.Store
	session      *mgo.Session
)

func initConfig() {
	var err error
	globalConfig, err = config.Load(configFile)
	if nil != err {
		log.WithError(err).Error("Could not load configurations from file - trying environment configurations instead.")

		globalConfig, err = config.LoadEnv()
		if nil != err {
			log.WithError(err).Error("Could not load configurations from environment")
		}
	}
}

// initializes the basic configuration for the log wrapper
func initLog() {
	err := globalConfig.Log.Apply()
	if nil != err {
		log.WithError(err).Panic("Could not apply logging configurations")
	}
}

// initializes distributed tracing
func initDistributedTracing() {
	log.Debug("Initializing distributed tracing")
	tracer, err := tracerfactory.Build(globalConfig.Tracing)
	if err != nil {
		log.WithError(err).Panic("Could not build a tracer")
	}

	opentracing.SetGlobalTracer(tracer)
}

// initializes the storage and managers
func initStorage() {
	log.WithField("dsn", globalConfig.Storage.DSN).Debug("Initializing storage")
	s, err := store.Build(globalConfig.Storage.DSN)
	if nil != err {
		log.Panic(err)
	}

	storage = s
}

// initializes the storage and managers
func initDatabase() {
	dsnURL, err := url.Parse(globalConfig.Database.DSN)
	switch dsnURL.Scheme {
	case "mongodb":
		log.Debug("MongoDB configuration chosen")

		log.WithField("dsn", globalConfig.Database.DSN).Debug("Trying to connect to MongoDB...")
		session, err = mgo.Dial(globalConfig.Database.DSN)
		if err != nil {
			log.Panic(err)
		}

		log.Debug("Connected to MongoDB")
		session.SetMode(mgo.Monotonic, true)
	default:
		log.Error("No Database selected")
	}
}
