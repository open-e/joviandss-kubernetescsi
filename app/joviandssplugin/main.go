package main

import (
	"flag"
	"fmt"
	"joviandss-kubernetescsi/pkg/common"
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
	addr       *string
	net        *string
	nodeId     *string
	driverName *string
	configPath *string
)

func main() {

	cfg := handleArgs()
	// TODO: check if logging parametrs a properly parse
	l := initLogging(cfg.LLevel, cfg.LDest)

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
		"obj": "Main",
	})

	return l
}

func handleArgs() *common.JovianDSSCfg {
	addr = flag.String("csi-address", "/var/lib/kubelet/plugins_registry/joviandss-csi-driver/csi.sock", "CSI endpoint socket address")
	net = flag.String("soc-type", "tcp", "CSI endpoint socket type")

	nodeId = flag.String("nodeid", "", "node id")
	configPath = flag.String("config", defaultConfigPath, "Path to configuration file")
	flag.Parse()

	if configPath == nil {
		fmt.Fprintf(os.Stderr, "No config file provided")
		os.Exit(1)
	}

	var cfg common.JovianDSSCfg
	if err := common.SetupConfig(*configPath, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to process config: %s", err.Error())
		os.Exit(1)
	}

	// if cfg.RestEndpoint.Addrs  != "" {
	// 	if *addr != defaultAddr {
	// 		cfg.Addr = *addr
	// 	}
	// } else {
	// 	cfg.Addr = *addr
	// }

	// if cfg.Network != "" {
	// 	if *addr != defaultAddr {
	// 		cfg.Network = *net
	// 	}
	// } else {
	// 	cfg.Network = *net
	// }

	// if len(*nodeId) > 0 {

	// 	cfg.Node.Id = cfg.Node.Id + *nodeId
	// }

	return &cfg
}

func routine(cfg *common.JovianDSSCfg, l *logrus.Entry) {
	l.Debug("Start app")
	//jdss, _ := joviandss.GetPlugin(cfg, l)

	//jdss.Run()
}
