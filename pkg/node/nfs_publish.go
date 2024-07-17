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

import "context"

import (
	"github.com/container-storage-interface/spec/lib/go/csi"

	mount "k8s.io/mount-utils"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

func (np *NodePlugin) NFSPublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) error {
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "PublishVolume",
		"proto":   "NFS",
		"section": "node",
	})

	l.Debugf("Publish Volume %+v", req.GetVolumeId())

	vcap := req.GetVolumeCapability()
	block := vcap.GetBlock() != nil

	if block {
		return status.Error(codes.Unimplemented, "Block attaching is not supported")
	}

	if mp, _ := np.mounter.IsMountPoint(req.GetTargetPath()); mp == true {
		return nil
	}
	mounter := np.mounter.(mount.SafeFormatAndMount)

	BindVolume(ctx, mounter, req.GetStagingTargetPath(), req.GetTargetPath(), req.GetReadonly())
	return nil
}

func (np *NodePlugin) NFSUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) error {
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "UnpublishVolume",
		"proto":   "NFS",
		"section": "node",
	})

	l.Debugf("Unpublish Volume %+v", req.GetVolumeId())
	mounter := np.mounter.(mount.MounterForceUnmounter)

	return UmountVolume(ctx, mounter, req.GetTargetPath())
}
