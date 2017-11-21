package config

import (
	"time"

	"github.com/hellofresh/janus/pkg/opentracing"
	"github.com/hellofresh/janus/pkg/provider/file"
	"github.com/hellofresh/janus/pkg/provider/mongodb"
	"github.com/hellofresh/janus/pkg/provider/web"
	"github.com/hellofresh/janus/pkg/types"
	"github.com/hellofresh/logging-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Specification for basic configurations
type Specification struct {
	Port                 int           `envconfig:"PORT"`
	Debug                bool          `envconfig:"DEBUG"`
	GraceTimeOut         int64         `envconfig:"GRACE_TIMEOUT"`
	MaxIdleConnsPerHost  int           `envconfig:"MAX_IDLE_CONNS_PER_HOST"`
	BackendFlushInterval time.Duration `envconfig:"BACKEND_FLUSH_INTERVAL"`
	CloseIdleConnsPeriod time.Duration `envconfig:"CLOSE_IDLE_CONNS_PERIOD"`
	Log                  logging.LogConfig
	Web                  *web.Provider
	File                 *file.Provider
	Mongodb              *mongodb.Provider
	Storage              Storage
	Stats                Stats
	Tracing              opentracing.Tracing
	TLS                  types.TLS
}

// Storage holds the configuration for a storage
type Storage struct {
	DSN string `envconfig:"STORAGE_DSN"`
}

// Stats holds the configuration for stats
type Stats struct {
	DSN                   string   `envconfig:"STATS_DSN"`
	Prefix                string   `envconfig:"STATS_PREFIX"`
	IDs                   string   `envconfig:"STATS_IDS"`
	AutoDiscoverThreshold uint     `envconfig:"STATS_AUTO_DISCOVER_THRESHOLD"`
	AutoDiscoverWhiteList []string `envconfig:"STATS_AUTO_DISCOVER_WHITE_LIST"`
	ErrorsSection         string   `envconfig:"STATS_ERRORS_SECTION"`
}

func init() {
	viper.SetDefault("port", "8080")
	viper.SetDefault("tls.port", "8433")
	viper.SetDefault("tls.redirect", true)
	viper.SetDefault("backendFlushInterval", "20ms")
	viper.SetDefault("database.dsn", "file:///etc/janus")
	viper.SetDefault("storage.dsn", "memory://localhost")
	viper.SetDefault("web.port", "8081")
	viper.SetDefault("web.tls.port", "8444")
	viper.SetDefault("web.tls.redisrect", true)
	viper.SetDefault("web.credentials.algorithm", "HS256")
	viper.SetDefault("web.credentials.basic.users", map[string]string{
		"admin": "admin",
	})
	viper.SetDefault("stats.dsn", "log://")
	viper.SetDefault("stats.errorsSection", "error-log")

	logging.InitDefaults(viper.GetViper(), "log")
}

//Load configuration variables
func Load(configFile string) (*Specification, error) {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("janus")
		viper.AddConfigPath("/etc/janus")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "No config file found")
	}

	var config Specification
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

//LoadEnv loads configuration from environment variables
func LoadEnv() (*Specification, error) {
	var config Specification

	// ensure the defaults are loaded
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
