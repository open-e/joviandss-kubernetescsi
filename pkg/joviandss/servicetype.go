package joviandss

import (
	"context"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

	ip *IdentityPlugin
	np *NodePlugin
	cp *ControllerPlugin
}

func GetPluginServer(cfg *Config, l *logrus.Entry) (s *PluginServer, err error) {
	s = &PluginServer{}

	s.l = l.WithFields(logrus.Fields{
		"node": cfg.NodeID,
		"obj":  "PluginServer",
	})

	if cfg.Network == "unix" {
		if err := os.Remove(cfg.Addr); err != nil && !os.IsNotExist(err) {
			s.l.Warn("Unable to clear unix socket %s. Error: %s", cfg.Addr, err)
			return nil, err
		}
	}

	listener, err := net.Listen(cfg.Network, cfg.Addr)
	s.listener = &listener
	if err != nil {
		s.l.Warnf("Unable to start listening socket %s %s. Error %s", cfg.Network, cfg.Addr, err)
		return nil, err
	}

	s.server = grpc.NewServer(grpc.UnaryInterceptor(s.grpcErrorHandler))

	for _, v := range cfg.Plugins {
		if v == IdentityPluginName {
			s.ip, err = GetIdentityPlugin(cfg, l)
			if err != nil {
				s.l.Warnf("Unable to create Identity Plugin: %s", err)
				continue
			}
			csi.RegisterIdentityServer(s.server, s.ip)
			s.l.Info("Register Identity Plugin")
		}

		if v == ControllerPluginName {
			s.cp, err = GetControllerPlugin(&cfg.Controller, l)

			if err != nil {
				s.l.Warnf("Unable to create Controller Plugin: %s", err)
				continue
			}
			csi.RegisterControllerServer(s.server, s.cp)
			s.l.Info("Register Controller Plugin")
		}

		if v == NodePluginName {
			s.np, err = GetNodePlugin(&cfg.Node, l)
			if err != nil {
				s.l.Warnf("Unable to create Node Plugin: %s", err)
				continue
			}
			csi.RegisterNodeServer(s.server, s.np)
			s.l.Info("Register Node Plugin")
		}
	}

	return s, nil

}

func (s *PluginServer) Run() (err error) {
	err = s.server.Serve(*s.listener)
	if err != nil {
		s.l.Warn("Unable to start listening on socket: %s", err)
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
		s.l.WithFields(logrus.Fields{"grpc": "Fail"}).Warn(err)
	}
	return resp, err
}
