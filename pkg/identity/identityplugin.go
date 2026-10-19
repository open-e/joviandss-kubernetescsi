package identity

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

type IdentityPlugin struct {
	l *log.Entry
}

func (ip *IdentityPlugin) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	l := ip.l.WithFields(log.Fields{
		"request": "GetPluginInfo",
		"func":    "GetPluginInfo",
		"section": "identity",
	})
	ctx = jcom.WithLogger(ctx, l)
	l.Debugf("Serving Plugin Info")

	if jcom.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          jcom.PluginName,
		VendorVersion: jcom.Version,
	}, nil
}

func GetIdentityPlugin(log *log.Entry) (ip *IdentityPlugin, err error) {
	ip = &IdentityPlugin{l: log}
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
