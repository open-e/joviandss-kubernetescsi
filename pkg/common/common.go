package common

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
	
	"joviandss-kubernetescsi/pkg/rest"

	"github.com/sirupsen/logrus"
)

// Version of plugin, should be filed during compilation
var Version string

// Plugin name
var PluginName = "joviandss-csi-iscsi.open-e.com"


type ISCSIEndpointCfg struct {
        Vnamelen         int
        Vpasslen         int
        Iqn              string
}

//ControllerCfg stores configaration properties of controller instance
type JovianDSSCfg struct {
	LLevel			string			`yaml:"loglevel"`
	LDest			string			`yaml:"logfile"`
	Pool			string			`yaml:"pool"`
        
	RestEndpointCfg		rest.RestEndpointCfg	`yaml:"endpoint"`
	ISCSIEndpointCfg	ISCSIEndpointCfg	`yaml:"iscsi"`
}


func GetLogger(logLevel string, toFile string) (*logrus.Logger, error) {
	log := logrus.New()

	formater := logrus.TextFormatter{

		DisableColors: false,
		FullTimestamp: true,
	}
	logrus.SetFormatter(&formater)

	if len(toFile) > 0 {
		file, err := os.OpenFile(toFile, os.O_CREATE|os.O_WRONLY, 0o640)
		if err == nil {
			log.Out = file
		} else {
			fmt.Fprintf(os.Stderr, "Logging to file error: %s\n", err.Error())
			return nil, err
		}
	} else {
		log.Out = os.Stdout
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogLevel processing error: %s\n", err.Error())
		return nil, err
	}

	log.SetLevel(lvl)
	
	return log, nil
}

// func SetupLogger(logLevel string, toFile string, l *logrus.Logger) (error)  {
// 
// 	formater := logrus.TextFormatter{
// 
// 		DisableColors: false,
// 		FullTimestamp: true,
// 	}
// 	l.SetFormatter(&formater)
// 
// 	if len(toFile) > 0 {
// 		file, err := os.OpenFile(toFile, os.O_CREATE|os.O_WRONLY, 0o640)
// 		if err == nil {
// 			l.Out = file
// 		} else {
// 			fmt.Fprintf(os.Stderr, "Logging to file error: %s\n", err.Error())
// 			return err
// 		}
// 	} else {
// 		l.Out = os.Stdout
// 	}
// 
// 	lvl, err := logrus.ParseLevel(logLevel)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "LogLevel processing error: %s\n", err.Error())
// 		return nil
// 	}
// 
// 	l.SetLevel(lvl)
// 	
// 	return nil
// }

//GetConfing reads Config from config file
func SetupConfig(path string, c *JovianDSSCfg) (error) {
        // var c JovianDSSCfg
        source, err := os.ReadFile(path)
        if err != nil {
                return err
        }

        err = yaml.Unmarshal(source, &c)
        if err != nil {
                return err
        }
        return nil
}
