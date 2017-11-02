package proxy

import (
	"encoding/json"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/types"
)

// Route is the container for a proxy and it's handlers
type Route struct {
	Proxy    *types.Proxy
	Inbound  InChain
	Outbound OutChain
}

type routeJSONProxy struct {
	Proxy *types.Proxy `json:"proxy"`
}

// NewRoute creates an instance of Route
func NewRoute(proxy *types.Proxy) *Route {
	return &Route{Proxy: proxy}
}

// NewRouteWithInOut creates an instance of Route with inbound and outbound handlers
func NewRouteWithInOut(proxy *types.Proxy, inbound InChain, outbound OutChain) *Route {
	return &Route{proxy, inbound, outbound}
}

// AddInbound adds inbound middlewares
func (r *Route) AddInbound(in ...router.Constructor) {
	for _, i := range in {
		r.Inbound = append(r.Inbound, i)
	}
}

// AddOutbound adds outbound middlewares
func (r *Route) AddOutbound(out ...OutLink) {
	for _, o := range out {
		r.Outbound = append(r.Outbound, o)
	}
}

// JSONMarshal encodes route struct to JSON
func (r *Route) JSONMarshal() ([]byte, error) {
	return json.Marshal(routeJSONProxy{r.Proxy})
}

// JSONUnmarshalRoute decodes route struct from JSON
func JSONUnmarshalRoute(rawRoute []byte) (*Route, error) {
	var proxyRoute routeJSONProxy
	if err := json.Unmarshal(rawRoute, &proxyRoute); err != nil {
		return nil, err
	}
	return NewRoute(proxyRoute.Proxy), nil
}

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
