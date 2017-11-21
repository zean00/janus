package opentracing

import (
	"context"
	"net/http"

	"github.com/hellofresh/gcloud-opentracing"
	"github.com/hellofresh/janus/pkg/opentracing/appdash"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

// GoogleCloudTracing holds the Google Application Default Credentials
type GoogleCloudTracing struct {
	ProjectID    string `envconfig:"TRACING_GC_PROJECT_ID"`
	Email        string `envconfig:"TRACING_GC_EMAIL"`
	PrivateKey   string `envconfig:"TRACING_GC_PRIVATE_KEY"`
	PrivateKeyID string `envconfig:"TRACING_GC_PRIVATE_ID"`
}

// AppdashTracing holds the Appdash tracing configuration
type AppdashTracing struct {
	DSN string `envconfig:"TRACING_APPDASH_DSN"`
	URL string `envconfig:"TRACING_APPDASH_URL"`
}

// Tracing represents the distributed tracing configuration
type Tracing struct {
	GoogleCloudTracing GoogleCloudTracing `mapstructure:"googleCloud"`
	AppdashTracing     AppdashTracing     `mapstructure:"appdash"`
}

// IsGoogleCloudEnabled checks if google cloud is enabled
func (t Tracing) IsGoogleCloudEnabled() bool {
	return len(t.GoogleCloudTracing.Email) > 0 && len(t.GoogleCloudTracing.PrivateKey) > 0 && len(t.GoogleCloudTracing.PrivateKeyID) > 0 && len(t.GoogleCloudTracing.ProjectID) > 0
}

// IsAppdashEnabled checks if appdash is enabled
func (t Tracing) IsAppdashEnabled() bool {
	return len(t.AppdashTracing.DSN) > 0
}

// Build a tracer based on the configuration provided
func Build(config Tracing) (opentracing.Tracer, error) {
	if config.IsGoogleCloudEnabled() {
		log.Debug("Using google cloud platform (stackdriver trace) as tracing system")

		tracer, err := gcloudtracer.NewTracer(
			context.Background(),
			gcloudtracer.WithLogger(log.StandardLogger()),
			gcloudtracer.WithProject(config.GoogleCloudTracing.ProjectID),
			gcloudtracer.WithJWTCredentials(gcloudtracer.JWTCredentials{
				Email:        config.GoogleCloudTracing.Email,
				PrivateKey:   []byte(config.GoogleCloudTracing.PrivateKey),
				PrivateKeyID: config.GoogleCloudTracing.PrivateKeyID,
			}),
		)
		if err != nil {
			return nil, err
		}

		return tracer, nil
	} else if config.IsAppdashEnabled() {
		server := appdash.NewServer(config.AppdashTracing.DSN, config.AppdashTracing.URL)

		appdashFields := log.WithFields(log.Fields{
			"appdash_dsn":    config.AppdashTracing.DSN,
			"appdash_web_ui": config.AppdashTracing.URL,
		})

		if config.AppdashTracing.URL != "" {
			appdashFields.Debug("Using local appdash server as tracing system")
			err := server.Listen()
			if err != nil {
				return nil, err
			}
		} else {
			appdashFields.Debug("Using remote appdash server as tracing system")
		}

		return server.GetTracer(), nil
	} else {
		log.Debug("No tracer selected")
		return &opentracing.NoopTracer{}, nil
	}
}

// FromContext creates a span from a context that contains a parent span
func FromContext(ctx context.Context, name string) opentracing.Span {
	span, _ := opentracing.StartSpanFromContext(ctx, name)
	return span
}

// ToContext sets a span to a context
func ToContext(r *http.Request, span opentracing.Span) *http.Request {
	return r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
}
