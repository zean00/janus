package oauth2

import (
	"fmt"

	"github.com/hellofresh/janus/pkg/jwt"
	"github.com/hellofresh/janus/pkg/notifier"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/server"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	repo        Repository
	adminRouter router.Router
)

func init() {
	plugin.RegisterEventHook(plugin.StartupEvent, onStartup)
	plugin.RegisterEventHook(plugin.AdminAPIStartupEvent, onAdminAPIStartup)
	plugin.RegisterPlugin("oauth2", plugin.Plugin{
		Action: setupOAuth2,
	})
}

// Config represents the oauth configuration
type Config struct {
	ServerName string `json:"server_name"`
}

func onAdminAPIStartup(event interface{}) error {
	e, ok := event.(plugin.OnAdminAPIStartup)
	if !ok {
		return errors.New("Could not convert event to admin startup type")
	}

	adminRouter = e.Router
	return nil
}

func onStartup(event interface{}) error {
	var (
		ntf notifier.Notifier
		err error
	)

	e, ok := event.(server.OnStartup)
	if !ok {
		return errors.New("Could not convert event to startup type")
	}

	if e.Config.Mongodb != nil {
		repo, err = NewMongoRepository(e.Config.Mongodb.Session)
	} else if e.Config.File != nil {
		authDir := fmt.Sprintf("%s/auth", e.Config.File.Directory)
		log.WithField("auth_dir", authDir).Debug("Trying to load configuration files")

		repo, err = NewFileSystemRepository(authDir)
		if err != nil {
			return errors.Wrap(err, "Could not create a file based repository for the oauth servers")
		}
	} else {
		return errors.New("The selected scheme is not supported to load OAuth servers")
	}

	if rawNtf := e.Notifier; rawNtf != nil {
		ntf = rawNtf.(notifier.Notifier)
	}

	loadOAuthEndpoints(adminRouter, repo, ntf, e.Config.Web.Credentials)
	loader := NewOAuthLoader(e.Register)
	loader.LoadDefinitions(repo)

	return nil
}

func setupOAuth2(route *proxy.Route, rawConfig plugin.Config) error {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return err
	}

	oauthServer, err := repo.FindByName(config.ServerName)
	if nil != err {
		return err
	}

	manager, err := getManager(oauthServer, config.ServerName)
	if nil != err {
		log.WithError(err).Error("OAuth Configuration for this API is incorrect, skipping...")
		return err
	}

	signingMethods, err := oauthServer.TokenStrategy.GetJWTSigningMethods()
	if err != nil {
		return err
	}

	route.AddInbound(NewKeyExistsMiddleware(manager))
	route.AddInbound(NewRevokeRulesMiddleware(jwt.NewParser(jwt.NewParserConfig(signingMethods...)), oauthServer.AccessRules))

	return nil
}

func getManager(oauthServer *OAuth, oAuthServerName string) (Manager, error) {
	managerType, err := ParseType(oauthServer.TokenStrategy.Name)
	if nil != err {
		return nil, err
	}

	return NewManagerFactory(oauthServer).Build(managerType)
}

// loadOAuthEndpoints register api endpoints
func loadOAuthEndpoints(router router.Router, repo Repository, ntf notifier.Notifier, cred types.Credentials) {
	log.Debug("Loading OAuth Endpoints")

	guard := jwt.NewGuard(cred)
	oAuthHandler := NewController(repo, ntf)
	oauthGroup := router.Group("/oauth/servers")
	oauthGroup.Use(jwt.NewMiddleware(guard).Handler)
	{
		oauthGroup.GET("/", oAuthHandler.Get())
		oauthGroup.GET("/{name}", oAuthHandler.GetBy())
		oauthGroup.POST("/", oAuthHandler.Post())
		oauthGroup.PUT("/{name}", oAuthHandler.PutBy())
		oauthGroup.DELETE("/{name}", oAuthHandler.DeleteBy())
	}
}
