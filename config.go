package main

import (
	"fmt"
	"io/ioutil"
	"net"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Verbose           bool              `yaml:"verbose" json:"verbose"`
	Interfaces        []string          `yaml:"interfaces" json:"interfaces"`
	ProxyAddress      string            `yaml:"proxy_address" json:"proxy_address"`
	AdminAddress      string            `yaml:"admin_address" json:"admin_address"`
	ProxyProxyConfigs ProxyProxyConfigs `yaml:"proxy_proxy_configs" json:"proxy_proxy_configs"`
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

	for name, ppc := range c.ProxyProxyConfigs {
		_, ipnet, err := net.ParseCIDR(ppc.InNet)
		if err != nil {
			errOut := fmt.Errorf("'in_net' (%s) of proxy_proxy_config %s could not be parsed as CIDR: %s\n", ppc.InNet, name, err)
			return c, errOut
		}

		c.ProxyProxyConfigs[name].inNet = ipnet
	}

	return c, nil
}

type ProxyProxyConfig struct {
	RemoteProxy string `yaml:"remote_proxy" json:"remote_proxy"`
	Verbose     bool   `json:"verbose" yaml:"verbose"`
	InNet       string `yaml:"in_net" json:"in_net"`

	inNet *net.IPNet `json:"-" yaml:"-"`
}

type ProxyProxyConfigs map[string]*ProxyProxyConfig

func (configs ProxyProxyConfigs) FindMatch(ips []net.IP) (string, *ProxyProxyConfig) {
	for name, config := range configs {
		for _, ip := range ips {
			if config.inNet.Contains(ip) {
				return name, config
			}
		}
	}

	return "none", &ProxyProxyConfig{}
}
