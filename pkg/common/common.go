package common

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Version of plugin, should be filed during compilation
var Version string

// Plugin name
var PluginName = "joviandss-csi-iscsi.open-e.com"


func GetLogger(logLevel string, toFile string) *logrus.Logger {
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
			os.Exit(1)
		}
	} else {
		log.Out = os.Stdout
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogLevel processing error: %s\n", err.Error())
		os.Exit(1)
	}

	log.SetLevel(lvl)
	
	return log
}

