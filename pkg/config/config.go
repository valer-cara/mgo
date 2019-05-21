package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/valer-cara/mgo/pkg/helm"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Helm struct {
		Repositories []helm.HelmRepo
	}
	Notification struct {
		Slack struct {
			Webhookurl string
			Channel    string
			Username   string
			Icon       string
		}
	}
}

var Global Config

func LoadGlobalConfig(path string) error {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(configFile, &Global); err != nil {
		return errors.New(fmt.Sprintf("YAML Unmarshal error: %s: %v", configFile, err))
	}

	return nil
}
