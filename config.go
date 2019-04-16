package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Verbose           bool              `yaml:"verbose"`
	MITMCert          string            `yaml:"mitm_cert"`
	MITMKey           string            `yaml:"mitm_key"`
	ProxyAddress      string            `yaml:"proxy_address"`
	AdminAddress      string            `yaml:"admin_address"`
	ProxyProxyConfigs ProxyProxyConfigs `yaml:"proxy_proxy_configs"`
}

func NewConfig(path string) (*Config, error) {
	c := &Config{}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		errOut := fmt.Errorf("Error while reading config file %s: %s\n", path, err)
		return c, errOut
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		errOut := fmt.Errorf("Error while unmarshalling config file %s: %s\n", path, err)
		return c, errOut
	}

	return c, nil
}

type ProxyProxyConfig struct {
	MITM        bool   `yaml:"mitm"`
	InNet       string `yaml:"in_net"`
	RemoteProxy string `yaml:"remote_proxy"`
}

type ProxyProxyConfigs map[string]ProxyProxyConfig

/*
func (configs ProxyProxyConfigs) FindMatch(ip net.IPAddr) (string, ProxyProxyConfig) {
	for name, config := range configs {
		if config.InNet.Contains(ip) {
			return name, config
		}
	}

	defaultConfig, ok := configs["default"]
	if ok {
		return "default", defaultConfig
	}

	return "none", ProxyProxyConfig{}
}
*/
