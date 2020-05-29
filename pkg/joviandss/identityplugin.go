package joviandss

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityPlugin struct {
	l   *logrus.Entry
	cfg *Config
}

func (ip *IdentityPlugin) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	ip.l.Trace("Serving GetPluginInfo")

	if ip.cfg.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if ip.cfg.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          ip.cfg.DriverName,
		VendorVersion: ip.cfg.Version,
	}, nil
}

func GetIdentityPlugin(conf *Config, log *logrus.Entry) (ip *IdentityPlugin, err error) {
	ip = &IdentityPlugin{cfg: conf, l: log}
	return ip, nil
}

func (ip *IdentityPlugin) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

func (ip *IdentityPlugin) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	ip.l.Infof("Using default capabilities")
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
