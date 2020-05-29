package joviandss

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/open-e/JovianDSS-KubernetesCSI/pkg/rest"
)

type ControllerCfg struct {
	Salt             string
	StorageEndpoints []rest.StorageCfg
	Vnamelen         int
	Vpasslen         int
	Nodeprefix       string
	Iqn              string
}

type NodeCfg struct {
	Id   string
	Addr string
	Port int
}

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

func GetConfing(path string) (*Config, error) {
	var c Config
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(source, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
