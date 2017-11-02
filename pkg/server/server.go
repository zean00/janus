package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/hellofresh/janus/pkg/config"
	"github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/metrics"
	"github.com/hellofresh/janus/pkg/middleware"
	"github.com/hellofresh/janus/pkg/notifier"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/provider"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/hellofresh/janus/pkg/web"
	stats "github.com/hellofresh/stats-go"
	log "github.com/sirupsen/logrus"
)

type (
	// Server is the reverse-proxy/load-balancer engine
	Server struct {
		globalConfiguration *config.Specification
		providers           []provider.Provider
		configurationChan   chan types.ConfigMessage
		stopChan            chan bool
		signals             chan os.Signal
		statsClient         stats.Client
		httpServer          *http.Server
		ntf                 notifier.Notifier
	}
)

var (
	errInvalidBackend    = errors.New(0, "API URI is invalid or not active, skipping...")
	errValidationBackend = errors.New(0, "validation errors")
)

// New returns a new instance of Server
func New(globalConfiguration *config.Specification) (*Server, error) {
	statsClient, err := metrics.NewStatsD(globalConfiguration.Stats)
	if err != nil {
		return nil, err
	}

	server := new(Server)
	server.statsClient = statsClient
	server.signals = make(chan os.Signal, 1)
	server.configurationChan = make(chan types.ConfigMessage, 100)
	server.stopChan = make(chan bool, 1)
	server.globalConfiguration = globalConfiguration
	server.configureSignals()

	return server, nil
}

// Start starts the server.
func (s *Server) Start() error {
	r := s.buildDefaultHTTPRouter()
	s.startServer(r)

	s.configureProviders()
	s.startProviders()
	go s.listenProviders()
	go s.listenSignals()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	defer log.Info("Server stopped")

	graceTimeOut := time.Duration(s.globalConfiguration.GraceTimeOut)
	ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
	log.Debugf("Waiting %s seconds before killing connections", graceTimeOut)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.WithError(err).Debug("Wait is over due to")
		s.httpServer.Close()
	}
	cancel()

	s.stopChan <- true
}

// Close closes the server
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.globalConfiguration.GraceTimeOut))
	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Warn("Timeout while stopping janus, killing instance ✝")
			os.Exit(1)
		}
	}(ctx)
	s.statsClient.Close()
	close(s.configurationChan)
	signal.Stop(s.signals)
	close(s.signals)
	close(s.stopChan)

	cancel()
}

// Wait blocks until server is shutted down.
func (s *Server) Wait() {
	<-s.stopChan
}

func (s *Server) configureProviders() {
	if s.globalConfiguration.File != nil {
		s.providers = append(s.providers, s.globalConfiguration.File)
	}
}

func (s *Server) startProviders() {
	// start providers
	for _, p := range s.providers {
		providerType := reflect.TypeOf(p)
		logger := log.WithField("provider_type", providerType)
		logger.Info("Starting provider")
		currentProvider := p
		go func() {
			err := currentProvider.Provide(s.configurationChan)
			if err != nil {
				logger.Error("Error starting provider")
			}
		}()
	}
}

func (s *Server) listenProviders() {
	for {
		select {
		case configMsg, ok := <-s.configurationChan:
			if !ok {
				return
			}

			newRouter := s.buildDefaultHTTPRouter()
			register := proxy.NewRegister(newRouter, proxy.Params{
				StatsClient:            s.statsClient,
				FlushInterval:          s.globalConfiguration.BackendFlushInterval,
				IdleConnectionsPerHost: s.globalConfiguration.MaxIdleConnsPerHost,
				CloseIdleConnsPeriod:   s.globalConfiguration.CloseIdleConnsPeriod,
			})

			s.configureBackends(register, configMsg.Configuration)

			s.httpServer.Handler = newRouter
		}
	}
}
func (s *Server) configureBackends(register *proxy.Register, backend []*types.Backend) {
	for _, backend := range backend {
		route, err := s.configureBackend(backend)
		if err != nil {
			log.WithError(err).Warn("Error ocurred when registering backend")
		}

		log.Debug("Backend registered")
		register.Add(route)
	}
}

func (s *Server) configureBackend(backend *types.Backend) (*proxy.Route, error) {
	logger := log.WithField("api_name", backend.Name)

	active, err := backend.Validate()
	if false == active && err != nil {
		return nil, errValidationBackend
	}

	if false == backend.Active {
		logger.Warn("Backend is not active, skipping...")
		active = false
	}

	if active {
		route := proxy.NewRoute(backend.Proxy)

		for _, pDefinition := range backend.Plugins {
			l := logger.WithField("name", pDefinition.Name)
			if pDefinition.Enabled {
				l.Debug("Plugin enabled")

				setup, err := plugin.DirectiveAction(pDefinition.Name)
				if err != nil {
					l.WithError(err).Error("Error loading plugin")
					continue
				}

				err = setup(route, pDefinition.Config)
				if err != nil {
					l.WithError(err).Error("Error executing plugin")
				}
			} else {
				l.Debug("Plugin not enabled")
			}
		}

		if len(backend.Proxy.Hosts) > 0 {
			route.AddInbound(middleware.NewHostMatcher(backend.Proxy.Hosts).Handler)
		}

		return route, nil
	}

	return nil, errInvalidBackend
}

func (s *Server) buildDefaultHTTPRouter() router.Router {
	// create router with a custom not found handler
	router.DefaultOptions.NotFoundHandler = errors.NotFound
	r := router.NewChiRouterWithOptions(router.DefaultOptions)
	r.Use(
		middleware.NewStats(s.statsClient).Handler,
		middleware.NewLogger().Handler,
		middleware.NewRecovery(errors.RecoveryHandler),
		middleware.NewOpenTracing(s.globalConfiguration.TLS.IsHTTPS()).Handler,
	)
	return r
}

func (s *Server) startServer(handler http.Handler) {
	address := fmt.Sprintf(":%v", s.globalConfiguration.Port)
	s.httpServer = &http.Server{Addr: address, Handler: handler}

	log.Info("Janus started")
	if s.globalConfiguration.TLS.IsHTTPS() {
		s.httpServer.Addr = fmt.Sprintf(":%v", s.globalConfiguration.TLS.Port)

		if s.globalConfiguration.TLS.Redirect {
			go func() {
				log.WithField("address", address).Info("Listening HTTP redirects to HTTPS")
				log.Fatal(http.ListenAndServe(address, web.RedirectHTTPS(s.globalConfiguration.TLS.Port)))
			}()
		}

		log.WithField("address", s.httpServer.Addr).Info("Listening HTTPS")
		s.httpServer.ListenAndServeTLS(s.globalConfiguration.TLS.CertFile, s.globalConfiguration.TLS.KeyFile)
	}

	log.WithField("address", address).Info("Certificate and certificate key were not found, defaulting to HTTP")
	s.httpServer.ListenAndServe()
}
