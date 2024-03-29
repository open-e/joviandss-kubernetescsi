package common

import (
	"fmt"
	"os"
	"encoding/base32"
	"strings"

	"gopkg.in/yaml.v3"
	
	// "joviandss-kubernetescsi/pkg/rest"
	uuid "github.com/google/uuid"

	"context"
	"github.com/sirupsen/logrus"
)

// Version of plugin, should be filed during compilation
var (
	Version string
	NodeID string
	LogLevel string
	LogPath string
)
// Plugin name
var PluginName = "iscsi.csi.joviandss.open-e.com"

var replacertojbase32 = strings.NewReplacer("=", "-")
var replacerfromjbase32 = strings.NewReplacer("-", "=")

// var replacertojbase64 = strings.NewReplacer("+", "_", "/", "-", "=", ".")
// var replacerfromjbase64 = strings.NewReplacer("_", "+", "-", "/", ".", "=")

var (
	NodeConfigPath string
	ControllerConfigPath string
)

type RestEndpointCfg struct {
	Addrs       []string		`json:"addrs,omitempty"`
	Port        int			`json:"port,omitempty"`
	Prot        string		`json:"prot,omitempty"`
	User        string		`json:"user,omitempty"`
	Pass        string		`json:"pass,omitempty"`
	IdleTimeOut string		`json:"idletimeout,omitempty"`
	Tries       int			`json:"tries,omitempty"`
}


type ISCSIEndpointCfg struct {
	Vnamelen        int		`json:"namelen,omitempty"`
	Vpasslen        int		`json:"passlen,omitempty"`
	Iqn		string		`json:"iqn,omitempty"`
	Addrs		[]string	`json:"addrs,omitempty"`
	Port		int		`json:"port,omitempty"`
}

//ControllerCfg stores configaration properties of controller instance
type JovianDSSCfg struct {
	LLevel			string			`yaml:"loglevel"`
	LDest			string			`yaml:"logfile"`
	Pool			string			`yaml:"pool"`
        
	RestEndpointCfg		RestEndpointCfg		`yaml:"endpoint"`
	ISCSIEndpointCfg	ISCSIEndpointCfg	`yaml:"iscsi"`
}


func GetLogger(logLevel string, toFile string) (*logrus.Logger, error) {
	log := logrus.New()
	
	// fmt.Printf("Getting logger with loglevel %s and logfile path %s", logLevel, toFile)

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
		fmt.Println("Print log to stdout")
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

type JDSSLoggerContextID int

const loggerKey JDSSLoggerContextID = iota

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {

	var traceId string

	id := ctx.Value("traceId")

	switch t := id.(type){

	case string:
		traceId = t
	case nil:
		traceId = uuid.Must(uuid.NewRandom()).String()
	default:
		traceId = uuid.Must(uuid.NewRandom()).String()
	}

	l := logger.WithFields(logrus.Fields{
		"traceId": traceId,
	})

	return context.WithValue(ctx, loggerKey, l)
}

// Logger From Context
func LFC(ctx context.Context) *logrus.Entry {
	
	l, ok := ctx.Value(loggerKey).(*logrus.Entry)

	if !ok {
		panic(fmt.Sprintf("Unable to get logger from context %+v", ctx))
	}

    return l
}

//Takes inut string and converts it to JBase64 string
func JBase32FromStr(in string) (out string) {
	out = base32.StdEncoding.EncodeToString([]byte(in))
	out = replacertojbase32.Replace(out)
	return out
}

//Takes JBase64 input and extracts original string
func JBase32ToStr(in string) (out string, err error) {
	out = replacerfromjbase32.Replace(in)
	bout, err := base32.StdEncoding.DecodeString(out)
	return string(bout[:]), err
}


func GetContext(traceId string) context.Context {
	ctxuuid := uuid.Must(uuid.NewRandom()).String()	
	ctx := context.Background() 
	return context.WithValue(ctx, "traceId", ctxuuid)
}
