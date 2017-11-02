package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hellofresh/janus/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Provider holds configurations of the provider.
type Provider struct {
	Directory string `envconfig:"PROVIDER_FILE_DIR" description:"Load configuration from one or more config files in a directory"`
}

// Provide allows the file provider to provide configurations to janus
// using the given configuration channel.
func (p *Provider) Provide(configChan chan<- types.ConfigMessage) error {
	backends, err := p.loadConfig()
	if err != nil {
		return err
	}

	sendConfigToChannel(configChan, backends)
	return nil
}

func (p *Provider) loadConfig() ([]*types.Backend, error) {
	if p.Directory != "" {
		return nil, errors.New("directory cannot be empty when you choose a file provider")
	}

	return loadFileConfigFromDirectory(fmt.Sprintf("%s/apis", p.Directory))
}

func sendConfigToChannel(configChan chan<- types.ConfigMessage, backends []*types.Backend) {
	configChan <- types.ConfigMessage{
		ProviderName:  "file",
		Configuration: backends,
	}
}

func loadFileConfigFromDirectory(dir string) ([]*types.Backend, error) {
	var backends []*types.Backend

	files, err := ioutil.ReadDir(dir)
	if nil != err {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name(), ".json") {
			filePath := filepath.Join(dir, f.Name())
			backend, err := loadFileConfig(filePath)
			if err != nil {
				log.WithError(err).WithField("path", filePath).Error("Couldn't load backend configuration")
			}
			backends = append(backends, backend)
		}
	}

	return backends, nil
}

func loadFileConfig(filePath string) (*types.Backend, error) {
	backendConfig, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not load backend form file")
	}

	backend := types.NewBackend()
	if err := json.Unmarshal(backendConfig, backend); err != nil {
		return nil, errors.Wrap(err, "could not decode backend configuration")
	}

	return backend, nil
}
