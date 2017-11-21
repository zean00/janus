package basic

import (
	"testing"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/server"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	route := proxy.NewRoute(&types.Proxy{})

	event1 := plugin.OnAdminAPIStartup{Router: router.NewChiRouter()}
	err := onAdminAPIStartup(event1)
	require.NoError(t, err)

	event2 := server.OnStartup{Register: proxy.NewRegister(router.NewChiRouter(), proxy.Params{})}
	err = onStartup(event2)
	require.NoError(t, err)

	err = setupBasicAuth(route, make(plugin.Config))
	require.NoError(t, err)
}

func TestOnStartupMissingAdminRouter(t *testing.T) {
	event := server.OnStartup{}
	err := onStartup(event)
	require.Error(t, err)
	require.IsType(t, ErrInvalidAdminRouter, err)
}

func TestOnStartupWrongEvent(t *testing.T) {
	wrongEvent := plugin.OnAdminAPIStartup{}
	err := onStartup(wrongEvent)
	require.Error(t, err)
}

func TestOnAdminAPIStartupWrongEvent(t *testing.T) {
	wrongEvent := server.OnStartup{}
	err := onAdminAPIStartup(wrongEvent)
	require.Error(t, err)
}
