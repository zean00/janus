package proxy

import (
	"testing"

	"github.com/hellofresh/janus/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRouteToJSON(t *testing.T) {
	definition := types.NewProxy()
	route := NewRoute(definition)
	json, err := route.JSONMarshal()
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"proxy": {"insecure_skip_verify": false, "append_path":false, "enable_load_balancing":false, "methods":[], "hosts":[], "preserve_host":false, "listen_path":"", "upstream_url":"", "strip_path":false, "upstreams": {"balancing": "", "targets": [] }}}`,
		string(json),
	)
}

func TestJSONToRoute(t *testing.T) {
	route, err := JSONUnmarshalRoute([]byte(`{"proxy": {"insecure_skip_verify": false, "append_path":false, "enable_load_balancing":false, "methods":[], "hosts":[], "preserve_host":false, "listen_path":"", "upstream_url":"/*", "strip_path":false}}`))

	assert.NoError(t, err)
	assert.IsType(t, &Route{}, route)
}

func TestJSONToRouteError(t *testing.T) {
	_, err := JSONUnmarshalRoute([]byte{})

	assert.Error(t, err)
}
