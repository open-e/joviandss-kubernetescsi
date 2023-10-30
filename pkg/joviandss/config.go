package joviandss

import (
	"os"

	"gopkg.in/yaml.v2"

	"JovianDSS-KubernetesCSI/pkg/rest"
)

// ControllerCfg stores configaration properties of controller instance
type ControllerCfg struct {
	Salt             string
	StorageEndpoints []rest.StorageCfg
	Vnamelen         int
	Vpasslen         int
	Nodeprefix       string
	Iqn              string
}

// NodeCfg storese info of node service
type NodeCfg struct {
	Id   string
	Addr string
	Port int
}

// Config stores config file representation
type Config struct {
	Salt    string
	Network string
	Addr    string

	LLevel string
	LDest  string

	Tries  int
	NodeID string

	DriverName string
	Version    string

	Plugins    []string
	Controller ControllerCfg
	Node       NodeCfg
}

// GetConfing reads Config from config file
func GetConfig(path string) (*Config, error) {
	var c Config
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(source, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
