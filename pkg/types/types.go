package types

import (
	"strings"

	"github.com/asaskevich/govalidator"
)

func init() {
	// initializes custom validators
	govalidator.CustomTypeTagMap.Set("urlpath", func(i interface{}, o interface{}) bool {
		s, ok := i.(string)
		if !ok {
			return false
		}

		return strings.Index(s, "/") == 0
	})
}

// Spec Holds an backend and basic options
type Spec struct {
	*Backend
}

// ConfigMessage hold configuration information exchanged between parts of janus.
type ConfigMessage struct {
	ProviderName  string
	Configuration []*Backend
}

// Plugin represents the plugins for an API
type Plugin struct {
	Name    string                 `bson:"name" json:"name"`
	Enabled bool                   `bson:"enabled" json:"enabled"`
	Config  map[string]interface{} `bson:"config" json:"config"`
}

// Backend holds backend configuration
type Backend struct {
	Name        string      `bson:"name" json:"name" valid:"required"`
	Active      bool        `bson:"active" json:"active"`
	Proxy       *Proxy      `bson:"proxy" json:"proxy" valid:"required"`
	Plugins     []Plugin    `bson:"plugins" json:"plugins"`
	HealthCheck HealthCheck `bson:"health_check" json:"health_check"`
}

// HealthCheck represents the health check configs
type HealthCheck struct {
	URL     string `bson:"url" json:"url" valid:"url"`
	Timeout int    `bson:"timeout" json:"timeout"`
}

// NewBackend creates a new backend with default values
func NewBackend() *Backend {
	return &Backend{
		Active:  true,
		Plugins: make([]Plugin, 0),
		Proxy:   NewProxy(),
	}
}

// Validate validates proxy data
func (d *Backend) Validate() (bool, error) {
	return govalidator.ValidateStruct(d)
}

// Proxy defines proxy rules for a route
type Proxy struct {
	ListenPath string `bson:"listen_path" json:"listen_path" mapstructure:"listen_path" valid:"required,urlpath"`
	// Deprecated: Use Upstreams instead.
	UpstreamURL        string     `bson:"upstream_url" json:"upstream_url" valid:"url"`
	Upstreams          *Upstreams `bson:"upstreams" json:"upstreams" mapstructure:"upstreams"`
	InsecureSkipVerify bool       `bson:"insecure_skip_verify" json:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
	StripPath          bool       `bson:"strip_path" json:"strip_path" mapstructure:"strip_path"`
	AppendPath         bool       `bson:"append_path" json:"append_path" mapstructure:"append_path"`
	PreserveHost       bool       `bson:"preserve_host" json:"preserve_host" mapstructure:"preserve_host"`
	Methods            []string   `bson:"methods" json:"methods"`
	Hosts              []string   `bson:"hosts" json:"hosts"`
}

// Upstreams represents a collection of targets where the requests will go to
type Upstreams struct {
	Balancing string    `bson:"balancing" json:"balancing"`
	Targets   []*Target `bson:"targets" json:"targets"`
}

// Target is an ip address/hostname with a port that identifies an instance of a backend service
type Target struct {
	Target string `bson:"target" json:"target" valid:"url,required"`
	Weight int    `bson:"weight" json:"weight"`
}

// NewProxy creates a new Proxy definition with default values
func NewProxy() *Proxy {
	return &Proxy{
		Methods: make([]string, 0),
		Hosts:   make([]string, 0),
		Upstreams: &Upstreams{
			Targets: make([]*Target, 0),
		},
	}
}

// Validate validates proxy data
func (d *Proxy) Validate() (bool, error) {
	return govalidator.ValidateStruct(d)
}
