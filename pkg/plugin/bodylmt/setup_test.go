package bodylmt

import (
	"testing"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	route := proxy.NewRoute(&types.Proxy{})
	err := setupBodyLimit(route, make(plugin.Config))
	assert.NoError(t, err)

	assert.Len(t, route.Inbound, 1)
}
