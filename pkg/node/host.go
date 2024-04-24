package node

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	mount "k8s.io/mount-utils"
	kexec "k8s.io/utils/exec"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

//var nodeId = ""

func GetNodeId(l *log.Entry) (string, error) {

	if len(jcom.NodeID) > 0 {
		l.Debugf("Node id identified %s", jcom.NodeID)

		return jcom.NodeID, nil
	}

	infostr := ""
	if out, err := exec.Command("hostname").Output(); err == nil {
		infostr = fmt.Sprintf("%s%s", infostr, out)
	}

	if out, err := exec.Command("cat", "/etc/machine-id").Output(); err == nil {
		infostr = fmt.Sprintf("%s%s", infostr, out)
	}

	if len(infostr) == 0 {
		return "", status.Errorf(codes.Internal, "Unable to identify node")
	}
	//l.Debugf("Node id %s", infostr)
	rawID := sha256.Sum256([]byte(infostr))
	jcom.NodeID = base64.StdEncoding.EncodeToString(rawID[:])

	//nodeId = string(rawID[:])

	return jcom.NodeID, nil
}

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

// FormatMountVolume tries to check fs on volume and formats if not sutable been found
func FormatMountVolume(ctx context.Context, volumeCapability csi.VolumeCapability, device string, location string) error {
	var err error
	var msg string
	m := mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      kexec.New()}

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "FormatMountVolume",
	})

	l.Debugf("Mounting %s to %s", device, location)
	if exists, err := mount.PathExists(location); exists == false {
		if err = os.MkdirAll(location, 0640); err != nil {
			msg = fmt.Sprintf("Unable to create directory %s, Error:%s", location, err.Error())
			return status.Error(codes.Internal, msg)
		}
		l.Debugf("Create dirrectory %s", location)
	}

	fsType := volumeCapability.GetMount().GetFsType()
	mOpt := volumeCapability.GetMount().GetMountFlags()

	if err = m.FormatAndMount(device, location, fsType, mOpt); err != nil {
		msg = fmt.Sprintf("Unable to mount device %s, Err: %s",
			location, err.Error())
		return status.Error(codes.Internal, msg)
	}

	l.Debugf("Mounting %s to %s done", device, location)

	return nil
}

// UnMountVolume unmounts volume
func UMountDevice(ctx context.Context, umounter mount.MounterForceUnmounter, device string) error {

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "UnMountVolume",
		"section": "node",
	})
	
	m := mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      kexec.New()}

	if mounts, err := m.GetMountRefs(device); err != nil {
		return err
	} else {
		for _, mpoint := range mounts {
			umounter.UnmountWithForce(mpoint, time.Minute)
		}
	}
	umounter.UnmountWithForce(device, time.Minute)
	return nil
}
