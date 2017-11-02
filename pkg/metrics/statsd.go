package metrics

import (
	"os"
	"path/filepath"

	"github.com/hellofresh/janus/pkg/config"
	stats "github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/bucket"
	"github.com/hellofresh/stats-go/hooks"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func NewStatsD(config config.Stats) (stats.Client, error) {
	sectionsTestsMap, err := bucket.ParseSectionsTestsMap(config.IDs)
	if err != nil {
		log.WithError(err).WithField("config", config.IDs).Error("Failed to parse stats second level IDs from env")
		sectionsTestsMap = map[bucket.PathSection]bucket.SectionTestDefinition{}
	}
	log.WithField("config", config.IDs).
		WithField("map", sectionsTestsMap.String()).
		Debug("Setting stats second level IDs")

	statsClient, err := stats.NewClient(config.DSN, config.Prefix)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing statsd client")
	}

	statsClient.SetHTTPMetricCallback(bucket.NewHasIDAtSecondLevelCallback(&bucket.SecondLevelIDConfig{
		HasIDAtSecondLevel:    sectionsTestsMap,
		AutoDiscoverThreshold: config.AutoDiscoverThreshold,
		AutoDiscoverWhiteList: config.AutoDiscoverWhiteList,
	}))

	host, err := os.Hostname()
	if nil != err {
		host = "-unknown-"
	}

	_, appFile := filepath.Split(os.Args[0])
	statsClient.TrackMetric("app", bucket.MetricOperation{"init", host, appFile})

	log.AddHook(hooks.NewLogrusHook(statsClient, config.ErrorsSection))

	return statsClient, nil
}
