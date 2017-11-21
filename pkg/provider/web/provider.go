package web

import (
	"fmt"
	"net/http"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/hellofresh/janus/pkg/checker"
	httpErrors "github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/jwt"
	"github.com/hellofresh/janus/pkg/middleware"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

// Provider holds configurations of the provider.
type Provider struct {
	Port                  int  `envconfig:"API_PORT"`
	ReadOnly              bool `envconfig:"API_READONLY"`
	Credentials           types.Credentials
	TLS                   types.TLS
	CurrentConfigurations *types.Configuration
}

// Provide allows the file provider to provide configurations to janus
// using the given configuration channel.
func (p *Provider) Provide(configChan chan<- types.ConfigMessage, configChangeChan chan types.ConfigurationEvent) error {
	log.Info("Janus Admin API starting...")
	router.DefaultOptions.NotFoundHandler = httpErrors.NotFound
	r := router.NewChiRouterWithOptions(router.DefaultOptions)

	// create authentication for Janus
	guard := jwt.NewGuard(p.Credentials)
	r.Use(
		chimiddleware.StripSlashes,
		chimiddleware.DefaultCompress,
		middleware.NewLogger().Handler,
		middleware.NewRecovery(httpErrors.RecoveryHandler),
		middleware.NewOpenTracing(p.TLS.IsHTTPS()).Handler,
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedHeaders:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowCredentials: true,
		}).Handler,
	)

	// create endpoints
	r.GET("/", Home())
	// health checks
	r.GET("/status", checker.NewOverviewHandler(p.CurrentConfigurations))
	r.GET("/status/{name}", checker.NewStatusHandler(p.CurrentConfigurations))

	handlers := jwt.Handler{Guard: guard}
	r.POST("/login", handlers.Login(p.Credentials))
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/refresh_token", handlers.Refresh())
	}

	p.loadAPIEndpoints(r, configChangeChan, guard)
	plugin.EmitEvent(plugin.AdminAPIStartupEvent, plugin.OnAdminAPIStartup{Router: r})

	go func() {
		p.listenAndServe(r)
	}()

	return nil
}

func (p *Provider) listenAndServe(handler http.Handler) error {
	address := fmt.Sprintf(":%v", p.Port)

	log.Info("Janus Admin API started")
	if p.TLS.IsHTTPS() {
		addressTLS := fmt.Sprintf(":%v", p.TLS.Port)
		if p.TLS.Redirect {
			go func() {
				log.WithField("address", address).Info("Listening HTTP redirects to HTTPS")
				log.Fatal(http.ListenAndServe(address, RedirectHTTPS(p.TLS.Port)))
			}()
		}

		log.WithField("address", addressTLS).Info("Listening HTTPS")
		return http.ListenAndServeTLS(addressTLS, p.TLS.CertFile, p.TLS.KeyFile, handler)
	}

	log.WithField("address", address).Info("Certificate and certificate key were not found, defaulting to HTTP")
	return http.ListenAndServe(address, handler)
}

//loadAPIEndpoints register api endpoints
func (p *Provider) loadAPIEndpoints(router router.Router, configChangeChan chan types.ConfigurationEvent, guard jwt.Guard) {
	log.Debug("Loading API Endpoints")

	// Apis endpoints
	handler := NewController(p.CurrentConfigurations)
	group := router.Group("/apis")
	group.Use(jwt.NewMiddleware(guard).Handler)
	{
		group.GET("/", handler.Get())
		group.GET("/{name}", handler.GetBy())

		if !p.ReadOnly {
			group.POST("/", handler.Post(configChangeChan))
			group.PUT("/{name}", handler.PutBy(configChangeChan))
			group.DELETE("/{name}", handler.DeleteBy(configChangeChan))
		}
	}
}
