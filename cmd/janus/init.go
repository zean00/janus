package main

import (
	"github.com/hellofresh/janus/pkg/config"
	tracerfactory "github.com/hellofresh/janus/pkg/opentracing"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

var (
	globalConfig *config.Specification
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
