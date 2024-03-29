package joviandss

import (
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

var supportedNodeServiceCapabilities = []csi.NodeServiceCapability_RPC_Type{

	csi.NodeServiceCapability_RPC_UNKNOWN,
	csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
}

// NodePlugin responsible for attaching and detaching volumes to host
type NodePlugin struct {
	cfg *NodeCfg
	l   *logrus.Entry
}

// GetNodePlugin inits NodePlugin
func GetNodePlugin(conf *NodeCfg, log *logrus.Entry) (np *NodePlugin, err error) {

	lFields := logrus.Fields{
		"node":   conf.Id,
		"plugin": "Node",
	}
	np = &NodePlugin{
		cfg: conf,
		l:   log.WithFields(lFields),
	}
	log.Debug(fmt.Sprintf("Config: %+v", *conf))
	return np, nil
}

// NodeExpandVolume responsible for update of file system on volume
func (np *NodePlugin) NodeExpandVolume(ctx context.Context, in *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	np.l.Trace("Expanding Volume")
	out := new(csi.NodeExpandVolumeResponse)
	return out, nil
}

// NodeGetInfo returns node info
func (np *NodePlugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {

	np.l.Tracef("NodeGetInfo: %+v", req)

	//TODO: Add node identification
	return &csi.NodeGetInfoResponse{
		NodeId: np.cfg.Id,
	}, nil
}

// NodeStageVolume introduce volume to host
func (np *NodePlugin) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {

	np.l.Tracef("Node Stage Volume")
	var msg string

	t, err := GetTargetFromReq(np.cfg, np.l, *req)
	if err != nil {
		return nil, err
	}
	var exists bool
	if exists, err = mount.PathExists(t.STPath); err != nil {
		msg = fmt.Sprintf("Unable to check file %s for volume %s. Err: %s", t.STPath, t.Tname, err.Error())
		t.l.Warn(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	// Some activity are taking place with target staging path
	if exists == false {
		if err = os.MkdirAll(t.STPath, 0640); err != nil {
			msg = fmt.Sprintf("Unable to create directory %s, Error:%s", t.TPath, err.Error())
			return nil, status.Error(codes.Internal, msg)

		}
	}

	// Volume do not exist
	err = t.SerializeTarget()
	if err != nil {
		return nil, err
	}

	err = t.StageVolume()

	if err != nil {
		t.DeleteSerialization()
		msg = fmt.Sprintf("Unable to stage volume: %s ", err.Error())
		np.l.Warn(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume remove volume from host
func (np *NodePlugin) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {
	// Log out from specified target
	var msg string
	np.l.Tracef("Node Unstage Volume %s", req.GetVolumeId())

	vname := req.GetVolumeId()
	if len(vname) == 0 {
		msg = fmt.Sprintf("Request do not contain volume id")
		np.l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	stp := req.GetStagingTargetPath()
	if len(stp) == 0 {
		msg = fmt.Sprintf("Request do not contain staging target path")
		np.l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if GetStageStatus(stp) == false {
		return &csi.NodeUnstageVolumeResponse{}, nil
	}
	t, err := GetTargetFromPath(np.cfg, np.l, stp)
	// TODO: implement recovery using target path
	if err != nil {
		msg = fmt.Sprintf("Unable to get info about target: %s", err.Error())
		np.l.Warn(msg)
		return nil, err
	}
	err = t.UnStageVolume()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	t.DeleteSerialization()
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume mount volume to target
func (np *NodePlugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	// TODO: ValidateCapability()

	np.l.Tracef("Node Publish Volume %s", req.GetVolumeId())

	block := false
	var msg string

	t, err := GetTargetFromReq(np.cfg, np.l, *req)
	if err != nil {
		return nil, err
	}

	if !block {
		err = t.FormatMountVolume(req)
	} else {
		return nil, status.Error(codes.Unimplemented, "Block attaching is not supported")
	}

	if err != nil {
		msg = fmt.Sprintf("Unable to mount volume: %s", err.Error())
		np.l.Warn(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmount volume
func (np *NodePlugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	np.l.Tracef("Node Unpublish Volume %s", req.GetVolumeId())

	block := false
	//eq := false
	var msg string

	tp := req.GetTargetPath()
	if len(tp) == 0 {
		msg = fmt.Sprintf("Request do not contain target path")
		np.l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	t, err := GetTarget(np.cfg, np.l, tp)
	if err != nil {
		return nil, err
	}

	if !block {
		err = t.UnMountVolume()
		if err != nil {
			msg = fmt.Sprintf("Unable to clean up on volume unmounting: %s", err.Error())
			return nil, status.Error(codes.Aborted, msg)
		}
	} else {
		return nil, status.Error(codes.Unimplemented, "Block detaching is not supported")
	}

	np.l.Tracef("Node Unpublish Volume %s Done.", req.GetVolumeId())

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetServiceCapability provides service capabilities
func NodeGetServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

// NodeGetCapabilities provides node capabilities
func (ns *NodePlugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	ns.l.Infof("Using default NodeGetCapabilities")

	var capabilities []*csi.NodeServiceCapability
	for _, c := range supportedNodeServiceCapabilities {
		capabilities = append(capabilities, NodeGetServiceCapability(c))
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil

}

// NodeGetVolumeStats volume total and available space
func (np *NodePlugin) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
