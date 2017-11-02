package loader

import (
	"testing"

	"net/http"

	"github.com/hellofresh/janus/pkg/api"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestLoadAPIsWithParams(t *testing.T) {
	r := router.NewChiRouter()
	Load(proxy.NewRegister(r, proxy.Params{}), api.NewInMemoryRepository())

	assert.Equal(t, 1, r.RoutesCount())
}

func TestLoadValidAPIDefinitions(t *testing.T) {
	r := router.NewChiRouter()

	apiRepo := api.NewInMemoryRepository()
	apiRepo.Add(&types.Backend{
		Name:   "test1",
		Active: true,
		Proxy: &types.Proxy{
			ListenPath:  "/test1",
			UpstreamURL: "http://test1",
			Methods:     []string{http.MethodGet},
		},
		Plugins: []api.Plugin{
			api.Plugin{
				Name:    "oauth2",
				Enabled: false,
			},
			api.Plugin{
				Name:    "compression",
				Enabled: true,
			},
			api.Plugin{
				Name:    "rate_limit",
				Enabled: true,
				Config:  map[string]interface{}{"limit": "10-S", "policy": "local"},
			},
		},
	})
	apiRepo.Add(&types.Backend{
		Name:   "test2",
		Active: true,
		Proxy: &types.Proxy{
			ListenPath:  "/test2",
			UpstreamURL: "http://test2",
			Methods:     []string{http.MethodGet},
		},
	})

	Load(proxy.NewRegister(r, proxy.Params{}), apiRepo)

	assert.Equal(t, 2, r.RoutesCount())
}

func TestLoadInvalidAPIDefinitions(t *testing.T) {
	r := router.NewChiRouter()

	apiRepo := api.NewInMemoryRepository()
	definition := &types.Backend{
		Name:   "test2",
		Active: true,
		Proxy: &types.Proxy{
			ListenPath:  "/test2",
			UpstreamURL: "http://test2",
			Methods:     []string{http.MethodGet},
		},
	}
	err := apiRepo.Add(definition)
	assert.NoError(t, err)

	definition.Name = ""
	Load(proxy.NewRegister(r, proxy.Params{}), apiRepo)

	assert.Equal(t, 1, r.RoutesCount())
}

func TestLoadAPIDefinitionsMissingHTTPMethods(t *testing.T) {
	r := router.NewChiRouter()

	apiRepo := api.NewInMemoryRepository()
	apiRepo.Add(&types.Backend{
		Name:   "test1",
		Active: true,
		Proxy: &types.Proxy{
			ListenPath:  "/test1",
			UpstreamURL: "http://test1",
		},
	})

	Load(proxy.NewRegister(r, proxy.Params{}), apiRepo)

	assert.Equal(t, 1, r.RoutesCount())
}

func TestLoadInactiveAPIDefinitions(t *testing.T) {
	r := router.NewChiRouter()

	apiRepo := api.NewInMemoryRepository()
	apiRepo.Add(&types.Backend{
		Name:   "test1",
		Active: false,
		Proxy: &types.Proxy{
			ListenPath:  "/test1",
			UpstreamURL: "http://test1",
		},
	})

	Load(proxy.NewRegister(r, proxy.Params{}), apiRepo)

	assert.Equal(t, 1, r.RoutesCount())
}
