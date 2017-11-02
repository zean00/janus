package main

import (
	"github.com/hellofresh/janus/pkg/server"

	"github.com/hellofresh/janus/pkg/notifier"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	// this is needed to call the init function on each plugin
	_ "github.com/hellofresh/janus/pkg/plugin/basic"
	_ "github.com/hellofresh/janus/pkg/plugin/bodylmt"
	_ "github.com/hellofresh/janus/pkg/plugin/compression"
	_ "github.com/hellofresh/janus/pkg/plugin/cors"
	_ "github.com/hellofresh/janus/pkg/plugin/oauth2"
	_ "github.com/hellofresh/janus/pkg/plugin/rate"
	_ "github.com/hellofresh/janus/pkg/plugin/requesttransformer"
	_ "github.com/hellofresh/janus/pkg/plugin/responsetransformer"

	// dynamically registered auth providers
	_ "github.com/hellofresh/janus/pkg/jwt/basic"
	_ "github.com/hellofresh/janus/pkg/jwt/github"

	// internal plugins
	_ "github.com/hellofresh/janus/pkg/loader"
	_ "github.com/hellofresh/janus/pkg/web"
)

// RunServer is the run command to start Janus
func RunServer(cmd *cobra.Command, args []string) {
	log.WithField("version", version).Info("Janus starting...")

	initConfig()
	initLog()
	initDistributedTracing()
	initStorage()
	initDatabase()

	defer globalConfig.Log.Flush()
	defer session.Close()

	if subscriber, ok := storage.(notifier.Subscriber); ok {
		listener := notifier.NewNotificationListener(subscriber)
		listener.Start(handleEvent)
	}

	if publisher, ok := storage.(notifier.Publisher); ok {
		ntf = notifier.NewPublisherNotifier(publisher, "")
	}

	svr, err := server.New(globalConfig)
	if err != nil {
		log.WithError(err).Error("Could not start the server")
	}
	svr.Start()
	defer svr.Close()

	svr.Wait()
	log.Info("Shutting down")
	log.Exit(0)

	event := plugin.OnStartup{
		Notifier:    ntf,
		StatsClient: statsClient,
		Register:    register,
		Config:      globalConfig,
	}
	plugin.EmitEvent(plugin.StartupEvent, event)
}

func handleEvent(notification notifier.Notification) {
	if notifier.RequireReload(notification.Command) {
		newRouter := createRouter()
		register := proxy.NewRegister(newRouter, proxy.Params{
			StatsClient:            statsClient,
			FlushInterval:          globalConfig.BackendFlushInterval,
			IdleConnectionsPerHost: globalConfig.MaxIdleConnsPerHost,
			CloseIdleConnsPeriod:   globalConfig.CloseIdleConnsPeriod,
		})

		plugin.EmitEvent(plugin.ReloadEvent, plugin.OnReload{Register: register})

		server.Handler = newRouter
	}
}
