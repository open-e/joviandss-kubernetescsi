/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
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
	"errors"
	"fmt"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	//"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
	kexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"

	jcom "joviandss-kubernetescsi/pkg/common"
)

const (
	deviceIPPath = "/host/dev/disk/by-path/ip"
)

// GetTarget constructs basic Target structure
func GetTarget(l *log.Entry, tp string) (t *Target, err error) {

	t = &Target{
		TPath:     tp,
		TProtocol: "tcp",
	}

	t.l = l

	l.Debug("Making Target")
	return t, nil
}

// GetTargetFromReq constructs Target structure from request data
func GetTargetFromReq(ctx context.Context, r interface{}) (t *Target, err error) {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "GetTargetFromReq",
		"section" : "target",
	})

	var pubContext map[string]string
	var msg string
	var vID string

	var fsType string
	var mountFlags []string

	sTPath := ""
	tPath := ""

	l.Debug("Processing request")
	if d, ok := r.(csi.NodeStageVolumeRequest); ok {

		l.Debug("Processing Stage request")
		pubContext = d.GetPublishContext()
		sTPath = d.GetStagingTargetPath()
		if len(sTPath) == 0 {
			msg = fmt.Sprintf("Request do not contain StagingTargetPath.")
			l.Warn(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}

		vID = d.GetVolumeId()
		if len(vID) == 0 {
			msg = fmt.Sprintf("Request do not contain volume id")
			l.Warn(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
		mount := d.GetVolumeCapability().GetMount()
		if mount != nil {
			fsType = mount.GetFsType()
			mountFlags = mount.GetMountFlags()
		}
	}

	if d, ok := r.(csi.NodePublishVolumeRequest); ok {

		l.Debug("Processing Publish request")

		pubContext = d.GetPublishContext()
		tPath = d.GetTargetPath()
		if len(tPath) == 0 {
			msg = fmt.Sprintf("Request do not contain TargetPath.")
			l.Warn(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}

		sTPath = d.GetStagingTargetPath()
		if len(sTPath) == 0 {
			msg = fmt.Sprintf("Request do not contain StagingTargetPath.")
			l.Warn(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}

		vID = d.GetVolumeId()
		if len(vID) == 0 {
			msg = fmt.Sprintf("Request do not contain volume id")
			l.Warn(msg)
			return nil, status.Error(codes.InvalidArgument, msg)
		}
	}

	var addrs []string
	if len(pubContext["addrs"]) > 0 {
		l.Debugf("addrs %s", pubContext["addrs"])
		addrs = strings.Split(pubContext["addrs"], ",")
		if len(addrs) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "Addrs are empty. No addresses provided.")
		}
	} else {
		l.Errorf("No JovianDSS address provideed in context %+v", pubContext)
		return nil, status.Errorf(codes.InvalidArgument, "Request context does not contain joviandss addresses")
	}

	var pp string
	if len(pubContext["port"]) > 0 {
		pp = pubContext["port"]
	} else {
		l.Debug("use default port: 3260")
		pp = "3260"
	}

	iqn := pubContext["iqn"]
	if len(iqn) == 0 {
		msg = fmt.Sprintf("Context do not contain iqn value")
		l.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	
	targetName := pubContext["target"]
	if len(targetName) == 0 {
		msg = fmt.Sprintf("Context do not contain target value")
		l.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	// coUser := ctx["name"]
	// if len(coUser) == 0 {
	// 	msg = fmt.Sprintf("Request do not contain CHAP name")
	// 	l.Warn(msg)
	// 	return nil, status.Error(codes.InvalidArgument, msg)
	// }

	// coPass := ctx["pass"]
	// if len(coUser) == 0 {
	// 	msg = fmt.Sprintf("Request do not contain CHAP pass")
	// 	l.Warn(msg)
	// 	return nil, status.Error(codes.InvalidArgument, msg)
	// }
	
	lun := pubContext["lun"]
	if len(lun) == 0 {
		l.Debug("Using default lun 0")
		lun = "0"
	}

	fullPortal := addrs[0] + ":" + pp

	dPath := strings.Join([]string{deviceIPPath, fullPortal, "iscsi", iqn, "lun", lun}, "-")

	// TODO: Provide default file system selection
	t = &Target{
		STPath:     sTPath,
		TPath:      tPath,
		DPath:      dPath,
		Portal:     addrs[0],
		PortalPort: pp,
		Iqn:        iqn,
		Tname:      targetName,
		Lun:        lun,
		//CoUser:     coUser, // Chap outgoing password
		//CoPass:     coPass, // Chap outgoing Password
		TProtocol:  "tcp",
		FsType:     "ext3",
		MountFlags: make([]string, 0),
	}

	l.Debugf("Target %+v", *t)
	if len(fsType) > 0 {
		t.FsType = fsType
	}

	if len(mountFlags) > 0 {
		t.MountFlags = mountFlags
	}

	t.l = l

	return t, nil
}

// GetTargetFromPath recoinstruct Target on the basis of the path
func GetTargetFromPath(l *log.Entry, path string) (t *Target, err error) {

	t = &Target{}
	tp := path + "/starget"
	err = t.DeSerializeTarget(tp)
	if err != nil {
		msg := fmt.Sprintf("Unable to serialize Target file %s. Error: %s", path, err.Error())
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	t.l = log.WithFields(log.Fields{
		"func": "GetTargetFromPath",
		"section": "node",
	})
	return t, nil
}

// SerializeTarget stores Target data to file
func (t *Target) SerializeTarget() error {

	var msg string
	d := *t
	//d.CoUser = "<Cleared>"
	//d.CoPass = "<Cleared>"

	data, err := yaml.Marshal(d)
	if err != nil {

		msg = fmt.Sprintf("Unable to serialize Target %+v.", d)
		return status.Error(codes.Internal, msg)
	}

	tp := t.STPath + "/starget"
	f, err := os.Create(tp)

	if err != nil {

		msg = fmt.Sprintf("Unable to create Target data file %s err %s", tp, err.Error())
		return status.Error(codes.Internal, msg)
	}

	defer f.Close()
	_, err = f.Write(data)

	if err != nil {
		msg = fmt.Sprintf("Unable to write Target data to %s err %s", tp, err.Error())
		return status.Error(codes.Internal, msg)
	}
	f.Sync()
	return nil
}

// DeSerializeTarget restores Target form data file
func (t *Target) DeSerializeTarget(stp string) error {
	var msg string

	data, err := os.ReadFile(stp)

	if err != nil {
		msg = fmt.Sprintf("Unable to read Target data file %s err %s", stp, err)
		return status.Error(codes.Internal, msg)
	}

	err = yaml.Unmarshal(data, t)
	if err != nil {
		msg = fmt.Sprintf("Unable to deirialize Target from file  %s", stp)
		t.l.Warn(msg)
		return status.Error(codes.Internal, msg)
	}

	return nil
}

// DeleteSerialization deletes record file about target
func (t *Target) DeleteSerialization() (err error) {
	var msg string
	stp := t.STPath + "/starget"
	var exists bool
	if exists, err = mount.PathExists(stp); err != nil {
		msg = fmt.Sprintf("Unable to identify serialization data for file %s. Because: %s", stp, err.Error())
		t.l.Warn(msg)
		return status.Error(codes.Internal, msg)

	}
	if exists == false {
		return nil
	}
	if err = os.Remove(t.STPath + "/starget"); err == nil {
		return nil
	}

	msg = fmt.Sprintf("Unable to delete serialized Target %s. Because: %s", stp, err.Error())
	t.l.Warn(msg)
	return err
}

// SetChapCred puts chap credantial to local db
// func (t *Target) SetChapCred() error {
// 
// 	tname := t.Iqn + ":" + t.Tname
// 
// 	t.l.Tracef("Target: %s", tname)
// 
// 	out, err := exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", tname, "-o", "update", "-n",
// 		"node.session.auth.authmethod", "-v", "CHAP").Output()
// 	if err != nil {
// 		t.l.Errorf("Could not update authentication method for %s error: %s", tname, string(out))
// 		return err
// 	}
// 
// 	out, err = exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", tname, "-o", "update", "-n",
// 		"node.session.auth.username", "-v", t.CoUser).Output()
// 	if err != nil {
// 		return fmt.Errorf("iscsi: failed to update node session user error: %v", string(out))
// 	}
// 	out, err = exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", tname, "-o", "update", "-n",
// 		"node.session.auth.password", "-v", t.CoPass).Output()
// 	if err != nil {
// 		return fmt.Errorf("iscsi: failed to update node session password error: %v", string(out))
// 	}
// 
// 	return nil
// }

// ClearChapCred sets chap credential to empty values
// func (t *Target) ClearChapCred() error {
// 
// 	tname := t.Iqn + ":" + t.Tname
// 
// 	portal := t.Portal + ":" + t.PortalPort
// 
// 	exec.Command("iscsiadm", "-m", "node", "-p", portal,
// 		"-T", tname, "-o", "update",
// 		"-n", "node.session.auth.password", "-v", "").Run()
// 	exec.Command("iscsiadm", "-m", "node", "-p", portal,
// 		"-T", tname, "-o", "update",
// 		"-n", "node.session.auth.username", "-v", "").Run()
// 
// 	return nil
// }

// FormatMountVolume tries to check fs on volume and formats if not sutable been found
func (t *Target) FormatMountVolume(req *csi.NodePublishVolumeRequest) error {
	var err error
	var msg string
	m := mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      kexec.New()}

	if exists, err := mount.PathExists(t.TPath); exists == false {
		if err = os.MkdirAll(t.TPath, 0640); err != nil {
			msg = fmt.Sprintf("Unable to create directory %s, Error:%s", t.TPath, err.Error())
			return status.Error(codes.Internal, msg)
		}
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	mOpt := req.GetVolumeCapability().GetMount().GetMountFlags()

	if err = m.FormatAndMount(t.DPath, t.TPath, fsType, mOpt); err != nil {
		msg = fmt.Sprintf("Unable to mount device %s, Err: %s",
			t.TPath, err.Error())
		return status.Error(codes.Internal, msg)
	}

	return nil
}

// UnMountVolume unmounts volume
func (t *Target) UnMountVolume(ctx context.Context) error {
	var err error
	var msg string

	var exists bool
	
	l := jcom.LFC(ctx)
	
	l = l.WithFields(log.Fields{
		"func": "UnMountVolume",
		"section": "node",
	})

	m := mount.New("")

	devices, mCount, err := mount.GetDeviceNameFromMount(m, t.TPath)
	if err != nil {
		msg = fmt.Sprintf("Unable to get device name from mount point %s, Err: %s", t.TPath, err.Error())
		t.l.Warn(msg)
		return status.Error(codes.Internal, msg)
	}

	if exists, err = mount.PathExists(t.TPath); err != nil {
		msg = fmt.Sprintf("Target path do not exists %s, Err: %s", t.TPath, err.Error())
		t.l.Warn(msg)
		return nil

	}

	if mCount == 0 && exists == false {
		t.l.Tracef("Target %s already umounted", t.TPath)
		return nil
	}

	if mCount > 0 {
		if err = m.Unmount(t.TPath); err != nil {
			msg = fmt.Sprintf("Unable to unmounted target %s for device %+v , Err: %s",
				t.TPath, devices, err.Error())
			t.l.Warn(msg)
			return status.Error(codes.Internal, msg)
		}
	}

	return mount.CleanupMountPoint(t.TPath, m, false)
}

// GetStageStatus check if specified dir exists
func GetStageStatus(stp string) bool {
	//TODO: check for presence of the device
	stp = stp + "/starget"
	if exists, _ := mount.PathExists(stp); exists == true {
		return true
	}

	return false
}

// StageVolume discovers iscsi target and attach it
func (t *Target) StageVolume(ctx context.Context) error {

	// Scan for targets
	l := jcom.LFC(ctx)
	
	l = l.WithFields(log.Fields{
		"func": "StageVolume",
		"section": "node",
	})

	//tname := t.Iqn // + ":" + t.Tname

	fullPortal := t.Portal + ":" + t.PortalPort

	devicePath := strings.Join([]string{deviceIPPath, fullPortal, "iscsi", t.Iqn, "lun", t.Lun}, "-")

	out, err := exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", t.Iqn, "-o", "new").Output()
	if err != nil {
		msg := fmt.Sprintf("Unable to add targetation %s error: %s", t.Iqn, err.Error())
		return errors.New(msg)
	}

	// Set properties

	// err = t.SetChapCred()
	// if err != nil {
	// 	msg := fmt.Sprintf("iscsi: failed to update iscsi node to portal %s error: %v", tname, err)
	// 	return errors.New(msg)
	// }

	//Attach Target
	// iscsiadm --mode discovery --op update --type sendtargets --portal targetIP
	// iscsiadm -m node -p 172.29.0.1 -T someiqn --login
	out, err = exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", t.Iqn, "--login").Output()
	if err != nil {
		//t.ClearChapCred()
		exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", t.Iqn, "--logout").Run()
		exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", t.Iqn, "-o", "delete").Run()
		msg := fmt.Sprintf("iscsi: failed to attach disk: Error: %s (%v)", string(out), err)
		return status.Errorf(codes.Internal, msg)
	}

	if exist := waitForPathToExist(&devicePath, 10, t.TProtocol); !exist {
		l.Errorf("Could not attach disk to the path %s: Timeout after 10s", devicePath)
		//t.ClearChapCred()
		exec.Command("iscsiadm", "-m", "node", "-p", t.Portal, "-T", t.Iqn, "-o", "delete").Run()
		msg := "Could not attach disk: Timeout after 10s"
		return status.Errorf(codes.Internal, msg)
	}

	return nil
}

// UnStageVolume detachs iscsi target from host
func (t *Target) UnStageVolume(ctx context.Context) error {

	// Scan for targets

	var msg string
	
	l := jcom.LFC(ctx)
	
	l = l.WithFields(log.Fields{
		"func": "StageVolume",
		"section": "node",
	})

	//tname := t.Iqn // + ":" + t.Tname

	portal := t.Portal + ":" + t.PortalPort

	if len(t.Iqn) == 0 {
		msg = fmt.Sprintf("Unable to get device target %s", t.Iqn)
		return errors.New(msg)
	}

	//err := t.ClearChapCred()
	//if err != nil {
	//	msg = fmt.Sprintf("Failed to clear ISCSI CHAP data %s error: %v", tname, err)
	//	return errors.New(msg)
	//}

	exec.Command("iscsiadm", "-m", "node", "-p", portal, "-T", t.Iqn, "--logout").Run()
	exec.Command("iscsiadm", "-m", "node", "-p", portal, "-T", t.Iqn, "-o", "delete").Run()

	return nil
}

type statFunc func(string) (os.FileInfo, error)
type globFunc func(string) ([]string, error)

func waitForPathToExist(devicePath *string, maxRetries int, deviceTransport string) bool {
	return waitForPathToExistInternal(devicePath, maxRetries, deviceTransport, os.Stat, filepath.Glob)
}

func waitForPathToExistInternal(devicePath *string, maxRetries int, deviceTransport string, osStat statFunc, filepathGlob globFunc) bool {
	if devicePath == nil {
		return false
	}

	for i := 0; i < maxRetries; i++ {
		var err error
		if deviceTransport == "tcp" {
			_, err = osStat(*devicePath)
		} else {
			fpath, _ := filepathGlob(*devicePath)
			if fpath == nil {
				err = os.ErrNotExist
			} else {
				*devicePath = fpath[0]
			}
		}
		if err == nil {
			return true
		}
		if !os.IsNotExist(err) {
			return false
		}
		if i == maxRetries-1 {
			break
		}
		time.Sleep(time.Second)
	}
	return false
}
