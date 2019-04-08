package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	log "github.com/sirupsen/logrus"
)

// Logger struct contains data and logic required for middleware functionality
type Logger struct{}

// NewLogger builds and returns new Logger middleware instance
func NewLogger() *Logger {
	return &Logger{}
}

// Handler implementation
func (m *Logger) Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"method": r.Method, "path": r.URL.Path}).Debug("Started request")

		fields := log.Fields{
			"request-id":  RequestIDFromContext(r.Context()),
			"method":      r.Method,
			"host":        r.Host,
			"request":     r.RequestURI,
			"remote-addr": r.RemoteAddr,
			"referer":     r.Referer(),
			"user-agent":  r.UserAgent(),
		}

		m := httpsnoop.CaptureMetrics(handler, w, r)

		fields["code"] = m.Code
		fields["duration"] = int(m.Duration / time.Millisecond)
		fields["duration-fmt"] = m.Duration.String()

		session := getSession(r)

		if session != nil {
			fields["session"] = session
		}

		log.WithFields(fields).Info("Completed handling request")
	})
}

func getSession(r *http.Request) map[string]interface{} {
	authHeaderValue := r.Header.Get("Authorization")
	parts := strings.Split(authHeaderValue, " ")
	if len(parts) < 2 {
		log.Warn("Attempted access with malformed header, no auth header found.")
		return nil
	}

	if strings.ToLower(parts[0]) != "bearer" {
		log.Warn("Bearer token malformed")
		return nil
	}
	accessToken := parts[1]

	tokenPart := strings.Split(accessToken, ".")
	if len(tokenPart) != 3 {
		log.Warn("token malformed")
		return nil
	}

	tokenClaim := tokenPart[1] + "=="

	b, err := base64.StdEncoding.DecodeString(tokenClaim)
	if err != nil {
		log.Warn("Claim malformed")
		return nil
	}

	var claim map[string]interface{}

	if err := json.Unmarshal(b, &claim); err != nil {
		log.Warn("Failed to unmarshal claim")
		return nil
	}

	return claim
}
