package main

import (
	"flag"
	"fmt"
	"joviandss-kubernetescsi/pkg/common"
	"joviandss-kubernetescsi/pkg/pluginserver"

	"os"

	"github.com/sirupsen/logrus"
	//"joviandss-kubernetescsi/pkg/joviandss"
)

func init() {
	flag.Set("logtostderr", "true")
}

const (
	defaultNetwork    = "unix"
	defaultAddr       = "/var/lib/kubelet/plugins_registry/com.open-e.joviandss.csi/csi.sock"
	defaultConfigPath = "/config/config.yaml"
)

var (
	driverName	*string
	address		string
	netType		string
	configPath	string
	logLevel	string
	logPath		string
	startController bool
	startNode	bool
	startIdentity	bool
)

func main() {

	cfg := handleArgs()

	// TODO: check if logging parametrs a properly parse
	var l *logrus.Entry
	if cfg != nil {
		l = initLogging(cfg.LLevel, cfg.LDest)
	} else {
		l = initLogging(logLevel, logPath)
	}

	routine(cfg, l)
	os.Exit(0)
}

func initLogging(logLevel string, toFile string) *logrus.Entry {
	log := logrus.New()
	formater := logrus.TextFormatter{

		DisableColors: false,
		FullTimestamp: true,
	}
	log.SetFormatter(&formater)

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

	l := log.WithFields(logrus.Fields{
		"section": "main",
	})

	return l
}

func handleArgs() *common.JovianDSSCfg {
	flag.StringVar(&address, "csi-address", "/var/lib/kubelet/plugins_registry/joviandss-csi-driver/csi.sock", "CSI endpoint socket address")
	flag.StringVar(&netType, "soc-type", "tcp", "CSI endpoint socket type")

	flag.BoolVar(&startController, "controller", false, "Start controller plugin")
	flag.BoolVar(&startNode, "node", false, "Start node plugin")
	flag.BoolVar(&startIdentity, "identity", false, "Start identity plugin")
	
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.StringVar(&logLevel, "loglevel", "WARNING", "Log Level, default is Warning")
	flag.StringVar(&logPath, "logpath", "/tmp/joviandsscsi", "Log file location")
	flag.Parse()

	if len(configPath) > 0 {
		var cfg common.JovianDSSCfg
		if err := common.SetupConfig(configPath, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to process config: %s", err.Error())
			os.Exit(1)
		}
		return &cfg
	}
	return nil
}

func routine(cfg *common.JovianDSSCfg, l *logrus.Entry) {
	l.Debug("Start app")
	jdss, _ := pluginserver.GetPluginServer(cfg, l, &netType, &address, startController, startNode, startIdentity)

	jdss.Run()
}
