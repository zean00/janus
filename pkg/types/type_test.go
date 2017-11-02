package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	instance := NewBackend()

	assert.IsType(t, &Backend{}, instance)
	assert.True(t, instance.Active)
}

func TestBackendSuccessfulValidation(t *testing.T) {
	instance := NewBackend()
	instance.Name = "Test"
	instance.Proxy.ListenPath = "/"
	instance.Proxy.UpstreamURL = "http://example.com"

	isValid, err := instance.Validate()
	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestBackendFailedValidation(t *testing.T) {
	instance := NewBackend()
	isValid, err := instance.Validate()

	assert.Error(t, err)
	assert.False(t, isValid)
}
func TestNewProxy(t *testing.T) {
	assert.NotNil(t, NewProxy())
}

func TestProxySuccessfulValidation(t *testing.T) {
	proxy := Proxy{
		ListenPath:  "/*",
		UpstreamURL: "http://test.com",
	}
	isValid, err := proxy.Validate()

	assert.NoError(t, err)
	assert.True(t, isValid)
}

func TestProxyEmptyListenPathValidation(t *testing.T) {
	proxy := Proxy{}
	isValid, err := proxy.Validate()

	assert.Error(t, err)
	assert.False(t, isValid)
}

func TestProxyInvalidTargetURLValidation(t *testing.T) {
	proxy := Proxy{
		ListenPath:  " ",
		UpstreamURL: "wrong",
	}
	isValid, err := proxy.Validate()

	assert.Error(t, err)
	assert.False(t, isValid)
}
