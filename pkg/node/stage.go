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
	"os/exec"
	"time"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mount "k8s.io/mount-utils"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// StageVolume discovers iscsi target and attachs it
func StageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) error {

	// Scan for targets
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "StageVolume",
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

	var port string
	if len(pubContext["port"]) > 0 {
		port = pubContext["port"]
	} else {
		l.Debug("use default port: 3260")
		port = "3260"
	}

	iqn := pubContext["iqn"]
	if len(iqn) == 0 {
		msg := fmt.Sprintf("Context do not contain iqn value")
		l.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	
	lun := pubContext["lun"]
	if len(lun) == 0 {
		l.Debug("Using default lun 0")
		lun = "0"
	}
	
	for _, addr := range addrs {

		fullPortal := addr + ":" + port

		_, err := exec.Command("iscsiadm", "-m", "discovery", "-p", fullPortal, "-t", iqn, "-o", "new").Output()
		if err != nil {
			l.Warnf("Unable to discover target record for target  %s error: %s", iqn, err.Error())
			continue
			// msg := fmt.Sprintf("Unable to discover target record for target  %s error: %s", iqn, err.Error())
			// return errors.New(msg)
		}

		_, err = exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "-o", "new").Output()
		if err != nil {
			l.Warnf("Unable to create target record for target  %s error: %s", iqn, err.Error())
			continue
			// msg := fmt.Sprintf("Unable to create target record for target  %s error: %s", iqn, err.Error())
			// return errors.New(msg)
		}

		out, err := exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "--login").Output()
		if err != nil {
			//t.ClearChapCred()
			exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "-o", "delete").Run()
			l.Warnf("Unable to togin into target %s error: %s (%v)", iqn, out, err.Error())
			continue
			// msg := fmt.Sprintf("Unable to togin into target %s error: %s (%v)", iqn, out, err.Error())
			// return status.Errorf(codes.Internal, msg)
		}

		devicePath := strings.Join([]string{deviceIPPath, fullPortal, "iscsi", iqn, "lun", lun}, "-")
		if err = FormatMountVolume(ctx, *req.GetVolumeCapability(), devicePath, req.GetStagingTargetPath()); err != nil {
			exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "--logout").Run()
			exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "-o", "delete").Run()
			msg := fmt.Sprintf("Failure in disk attaching procedure, err %s", err.Error())
			l.Error(msg)
			return status.Errorf(codes.Internal, msg)
		}
		return nil
	}

	return nil
}


// StageVolume discovers iscsi target and attachs it
func (np *NodePlugin)UnStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) error {
	var msg string

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "StageVolume",
		"section": "node",
	})
	umounter := np.mounter.(mount.MounterForceUnmounter)
	if mp, _ := np.mounter.IsMountPoint(req.GetStagingTargetPath()); mp == true {
		if err := umounter.UnmountWithForce(req.GetStagingTargetPath() , time.Minute); err != nil {
			msg = fmt.Sprintf("Failure in umounting %s unmounting %s", req.GetStagingTargetPath(), err.Error())
			l.Warn(msg)
			return status.Error(codes.Internal, msg)
		}
	}
	var pubContext = req.GetPublishContext()

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

	var port string

	
	if len(pubContext["port"]) > 0 {
		port = pubContext["port"]
	} else {
		l.Debug("use default port: 3260")
		port = "3260"
	}

	iqn := pubContext["iqn"]
	if len(iqn) == 0 {
		msg := fmt.Sprintf("Context do not contain iqn value")
		l.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	lun := pubContext["lun"]
	if len(lun) == 0 {
		l.Debug("Using default lun 0")
		lun = "0"
	}

	for _, addr := range addrs {
		fullPortal := addr + ":" + port

		devicePath := strings.Join([]string{deviceIPPath, fullPortal, "iscsi", iqn, "lun", lun}, "-")

		if exists, _ := mount.PathExists(devicePath); exists {

			UMountDevice(ctx, umounter, devicePath)
			exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "--logout").Run()
			exec.Command("iscsiadm", "-m", "node", "-p", fullPortal, "-T", iqn, "-o", "delete").Run()
			return nil
		}
	}

	return nil

}
