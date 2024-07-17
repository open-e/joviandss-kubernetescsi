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

package pluginserver

import (
	"context"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jcntr "github.com/open-e/joviandss-kubernetescsi/pkg/controller"
	jidnt "github.com/open-e/joviandss-kubernetescsi/pkg/identity"
	jnode "github.com/open-e/joviandss-kubernetescsi/pkg/node"
)

const (
	ControllerPluginName = "CONTROLLER_SERVICE"
	IdentityPluginName   = "IDENTITY_SERVICE"
	NodePluginName       = "NODE_SERVICE"
)

type PluginServer struct {
	server   *grpc.Server
	listener *net.Listener
	l        *logrus.Entry
}

func GetPluginServer(cfg *common.JovianDSSCfg, l *logrus.Entry, netType *string, addr *string, cntrSrv bool, nodeSrv bool, identitySrv bool) (s *PluginServer, err error) {
	s = &PluginServer{}

	l = l.WithFields(logrus.Fields{
		"func":    "GetPluginServer",
		"section": "PluginServer",
	})
	s.l = l
	if *netType == "unix" {
		if err := os.Remove(*addr); err != nil && !os.IsNotExist(err) {
			s.l.Warnf("Unable to clear unix socket %s. Error: %s", *addr, err)
			return nil, err
		}
	}

	listener, err := net.Listen(*netType, *addr)
	s.listener = &listener
	if err != nil {
		s.l.Warnf("Unable to start listening socket %s %s. Error %s", *netType, *addr, err)
		return nil, err
	}

	s.server = grpc.NewServer(grpc.UnaryInterceptor(s.grpcErrorHandler), grpc.MaxConcurrentStreams(128))

	if identitySrv {
		ip, err := jidnt.GetIdentityPlugin(l)
		if err != nil {
			l.Warnf("Unable to setup Identity Plugin: %s", err)
		}
		csi.RegisterIdentityServer(s.server, ip)
		l.Info("Register Identity Plugin")
	}

	if cntrSrv {
		var cp jcntr.ControllerPlugin

		if err = jcntr.SetupControllerPlugin(&cp, cfg); err == nil {
			l.Info("Register Controller Plugin")

			csi.RegisterControllerServer(s.server, &cp)

		} else {
			l.Warnf("Unable to create Controller Plugin: %s", err)
			return nil, err
		}

	}

	if nodeSrv {
		if np, err := jnode.GetNodePlugin(l); err != nil {
			l.Warnf("Unable to create Node Plugin: %s", err.Error())
			return nil, err
		} else {
			l.Debug("Register Node Plugin")

			csi.RegisterNodeServer(s.server, np)
		}
	}

	return s, nil
}

func (s *PluginServer) Run() (err error) {
	err = s.server.Serve(*s.listener)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"func": "Run",
		}).Warnf("Unable to start listening on socket: %s", err)
		return err
	}
	return nil
}

func (s *PluginServer) grpcErrorHandler(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"func": "grpcErrorhandler",
		}).Warn(err.Error())
	}
	return resp, err
}
