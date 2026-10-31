/*
Copyright (c) 2024 Open-E, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package common

import (
	"context"
	"encoding/base32"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// "joviandss-kubernetescsi/pkg/rest"
	uuid "github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

var (
	replacertojbase32   = strings.NewReplacer("=", "-")
	replacerfromjbase32 = strings.NewReplacer("-", "=")
)

type StorageAccessProtocolType string

const (
	ISCSI StorageAccessProtocolType = "iscsi"
	NFS   StorageAccessProtocolType = "nfs"
)

type RestEndpointCfg struct {
	Addrs       []string `json:"addrs,omitempty"`
	Port        int      `json:"port,omitempty"`
	Prot        string   `json:"prot,omitempty"`
	User        string   `json:"user,omitempty"`
	Pass        string   `json:"pass,omitempty"`
	IdleTimeOut string   `json:"idletimeout,omitempty"`
	Tries       int      `json:"tries,omitempty"`
}

type ISCSIEndpointCfg struct {
	Vnamelen int      `json:"namelen,omitempty"`
	Vpasslen int      `json:"passlen,omitempty"`
	Iqn      string   `json:"iqn,omitempty"`
	Addrs    []string `json:"addrs,omitempty"`
	Port     int      `json:"port,omitempty"`
}

type NFSEndpointCfg struct {
	Addrs []string `json:"addrs,omitempty"`
}

// ControllerCfg stores configaration properties of controller instance
type JovianDSSCfg struct {
	LLevel string `yaml:"loglevel"`
	LDest  string `yaml:"logfile"`
	Pool   string `yaml:"pool"`

	RestEndpointCfg  RestEndpointCfg   `yaml:"endpoint"`
	ISCSIEndpointCfg *ISCSIEndpointCfg `yaml:"iscsi"`
	NFSEndpointCfg   *NFSEndpointCfg   `yaml:"nfs"`
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

// GetConfing reads Config from config file
func SetupConfig(path string, c *JovianDSSCfg) error {
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

	switch t := id.(type) {

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

// Takes inut string and converts it to JBase64 string
func JBase32FromStr(in string) (out string) {
	out = base32.StdEncoding.EncodeToString([]byte(in))
	out = replacertojbase32.Replace(out)
	return out
}

// Takes JBase64 input and extracts original string
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

func InitVars() error {
	PluginProtocol = StorageAccessProtocolType(PluginProtocolCompileString)
	if PluginProtocol == ISCSI {
		PluginName = "iscsi.csi.joviandss.open-e.com"
	} else if PluginProtocol == NFS {
		PluginName = "nfs.csi.joviandss.open-e.com"
	} else {
		return status.Errorf(codes.InvalidArgument, "Unable to identify driver type")
	}
	return nil
}

func init() {
	PluginProtocol = StorageAccessProtocolType(PluginProtocolCompileString)
	if PluginProtocol == ISCSI {
		PluginName = "iscsi.csi.joviandss.open-e.com"
	} else if PluginProtocol == NFS {
		PluginName = "nfs.csi.joviandss.open-e.com"
	}
}
