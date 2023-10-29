package controller

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
	
	"github.com/open-e/JovianDSS-KubernetesCSI/pkg/rest"
)

//ControllerCfg stores configaration properties of controller instance
type ControllerCfg struct {
	StorageEndpoints []rest.StorageCfg
	Vnamelen         int
	Vpasslen         int
	Iqn              string
	
	LLevel		 string `yaml:"llevel"`
	LPath		 string `yaml:"lpath"`
}

func GetConfing(path string, cfg *ControllerCfg) error {

	//cfg.LPath = "/var/log/joviandss-csi-controller"
	if source, err := os.ReadFile(path); err != nil {
		return nil
	} else {
		if uerr := yaml.Unmarshal(source, cfg); uerr != nil {
			log.Fatalf("Controller config file processing error: %v", uerr)
		}
	}
	return nil
}
