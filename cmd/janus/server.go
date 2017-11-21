package main

import (
	"github.com/hellofresh/janus/pkg/server"

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
)

// RunServer is the run command to start Janus
func RunServer(cmd *cobra.Command, args []string) {
	log.WithField("version", version).Info("Janus starting...")

	initConfig()
	initLog()
	initDistributedTracing()

	defer globalConfig.Log.Flush()

	svr, err := server.New(globalConfig)
	if err != nil {
		log.WithError(err).Error("Could not start the server")
	}
	svr.Start()
	defer svr.Close()

	svr.Wait()
	log.Info("Shutting down")
	log.Exit(0)
}
