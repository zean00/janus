package oauth2

import (
	"net/http"
	"strings"

	"github.com/hellofresh/janus/pkg/jwt"
	log "github.com/sirupsen/logrus"
)

// NewRevokeRulesMiddleware creates a new revoke rules middleware
func NewRevokeRulesMiddleware(parser *jwt.Parser, accessRules []*AccessRule) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.WithField("rules", len(accessRules)).Debug("Starting revoke rules middleware")

			// If no rules are set then lets not parse the token to avoid performance issues
			/*
				if len(accessRules) < 1 {
					handler.ServeHTTP(w, r)
					return
				}
			*/

			token, err := parser.ParseFromRequest(r)
			if err != nil {
				log.WithError(err).Debug("Could not parse the JWT")
				handler.ServeHTTP(w, r)
				return
			}

			if claims, ok := parser.GetMapClaims(token); ok && token.Valid {
				for k, v := range claims {
					val, ok := v.(string)
					if !ok {
						continue
					}
					if strings.ToLower(k) == "sub" {
						r.Header.Set("subject", val)
					} else if strings.ToLower(k) == "aud" {
						r.Header.Set("audience", val)
					} else if strings.ToLower(k) == "iss" {
						r.Header.Set("issuer", val)
					} else {
						r.Header.Set(k, v.(string))
					}
				}

				if len(accessRules) < 1 {
					handler.ServeHTTP(w, r)
					return
				}

				for _, rule := range accessRules {
					allowed, err := rule.IsAllowed(claims)
					if err != nil {
						log.WithError(err).Debug("Rule is not allowed")
						continue
					}

					if allowed {
						handler.ServeHTTP(w, r)
					} else {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
				}
			}

			handler.ServeHTTP(w, r)
		})
	}
}
