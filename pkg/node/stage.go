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
	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// StageVolume discovers iscsi target and attachs it
func (np *NodePlugin) StageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) error {
	// Scan for targets
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "StageVolume",
		"section": "node",
	})

	pubContext := req.GetPublishContext()

	protocol := pubContext["protocol_type"]

	if len(protocol) > 0 && protocol == "NFS" {
		return np.NFSStageVolume(ctx, req)
	} else {
		return np.ISCSiStageVolume(ctx, req)
	}
}

// StageVolume discovers iscsi target and attachs it
func (np *NodePlugin) UnStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) error {
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "UnStageVolume",
		"section": "node",
	})
	pubContext := req.GetPublishContext()

	protocol := pubContext["protocol_type"]

	if len(protocol) > 0 && protocol == "NFS" {
		return np.NFSUnStageVolume(ctx, req)
	} else {
		return np.ISCSiUnStageVolume(ctx, req)
	}
}
