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

package node

import (
	"fmt"
	"strings"
	// "time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mount "k8s.io/mount-utils"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// StageVolume discovers iscsi target and attachs it
func (np *NodePlugin) NFSStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (err error) {
	// Scan for targets
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "StageVolume",
		"proto":   "NFS",
		"section": "node",
	})

	pubContext := req.GetPublishContext()

	var addrs []string
	if len(pubContext["addrs"]) > 0 {
		l.Debugf("addrs %s", pubContext["addrs"])
		addrs = strings.Split(pubContext["addrs"], ",")
		if len(addrs) == 0 {
			return status.Errorf(codes.InvalidArgument, "Addrs are empty. No addresses provided.")
		}
	} else {
		l.Errorf("No JovianDSS address provideed in context %+v", pubContext)
		return status.Errorf(codes.InvalidArgument, "Request context does not contain joviandss addresses")
	}

	sharePath := pubContext["share_path"]
	if len(sharePath) == 0 {
		msg := fmt.Sprintf("Context do not contain share_path value")
		l.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	for _, addr := range addrs {

		if err = MountNFSVolume(ctx, *req.GetVolumeCapability(), addr, sharePath, req.GetStagingTargetPath()); err != nil {
			l.Warn(err)
			continue
			// return status.Errorf(codes.Internal, msg)
		}
		return nil
	}

	return err
}

// StageVolume discovers iscsi target and attachs it
func (np *NodePlugin) NFSUnStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) error {
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "StageVolume",
		"proto":   "NFS",
		"section": "node",
	})
	l.Debugf("Unstaging volume %s", req.GetVolumeId())
	var ok bool
	var umounter mount.MounterForceUnmounter
	if umounter, ok = np.mounter.(mount.MounterForceUnmounter); ok {
		if err := UmountVolume(ctx, umounter, req.GetStagingTargetPath()); err != nil {
			return err
		}
	} else {
		return status.Error(codes.Internal, "Unable to identify umounter")
	}

	pubContext := req.GetPublishContext()
	var addrs []string
	if len(pubContext["addrs"]) > 0 {
		l.Debugf("addrs %s", pubContext["addrs"])
		addrs = strings.Split(pubContext["addrs"], ",")
		if len(addrs) == 0 {
			return status.Errorf(codes.InvalidArgument, "Addrs are empty. No addresses provided.")
		}
	} else {
		l.Errorf("No JovianDSS address provideed in context %+v", pubContext)
		return status.Errorf(codes.InvalidArgument, "Request context does not contain joviandss addresses")
	}

	sharePath := pubContext["share_path"]

	if len(sharePath) == 0 {
		msg := fmt.Sprintf("Context do not contain share_path value")
		l.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	for _, addr := range addrs {
		mnt := fmt.Sprintf("%s:%s", addr, sharePath)
		if err := UmountVolume(ctx, umounter, mnt); err != nil {
			return err
		}
	}
	return nil
}
