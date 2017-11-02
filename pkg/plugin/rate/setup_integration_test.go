// +build integration

package rate

import (
	"testing"

	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitPluginRedisPolicy(t *testing.T) {
	rawConfig := map[string]interface{}{
		"limit":  "10-S",
		"policy": "redis",
		"redis": map[string]interface{}{
			"dsn":    "localhost",
			"prefix": "test",
		},
	}

	route := proxy.NewRoute(&types.Proxy{})
	err := setupRateLimit(route, rawConfig)

	assert.Error(t, err)
}
