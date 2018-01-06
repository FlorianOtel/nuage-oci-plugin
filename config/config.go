package config

import (
	"io/ioutil"

	nuagecni "github.com/OpenPlatformSDN/nuage-cni/config"

	yaml "gopkg.in/yaml.v2"
)

// nuage-oci-plugin -- configuration file
type Config struct {
	// Not supplied in YAML config file
	ConfigFile string `yaml:"-"`
	// Config file fields
	AgentServerConfig nuagecni.AgentConfig `yaml:"agent-config"`
}

func LoadConfig(conf *Config) error {
	data, err := ioutil.ReadFile(conf.ConfigFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	return nil
}
