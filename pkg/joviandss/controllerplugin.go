package joviandss

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/open-e/JovianDSS-KubernetesCSI/pkg/rest"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

const (
	minVolumeSize = 16 * mib
)

var supportedControllerCapabilities = []csi.ControllerServiceCapability_RPC_Type{

	csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
	csi.ControllerServiceCapability_RPC_GET_CAPACITY,

	//TODO:
	//csi.ControllerServiceCapability_RPC_PUBLISH_READONLY,
}

var supportedVolumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
	//VolumeCapability_AccessMode_UNKNOWN,
	csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
	//VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
	//VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	//VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,

}

// ControllerPlugin provides CSI controller plugin interface
type ControllerPlugin struct {
	l                *logrus.Entry
	cfg              *ControllerCfg
	iqn              string
	snapReg          string
	volumesAccess    sync.Mutex
	volumesInProcess map[string]bool

	endpoints    []*rest.StorageInterface
	capabilities []*csi.ControllerServiceCapability
	vCap         []*csi.VolumeCapability
}

type origin struct {
	Pool     string
	Volume   string
	Snapshot string
}

func parseOrigin(or string) (*origin, error) {
	var out origin
	poolAndName := strings.Split(or, "/")

	if len(poolAndName) != 2 {
		msg := fmt.Sprintf("Incorrecct origin %s", or)
		return nil, status.Errorf(codes.Internal, msg)
	}

	out.Pool = poolAndName[0]
	nameAndSnap := strings.Split(poolAndName[1], "@")
	if len(poolAndName) != 2 {
		msg := fmt.Sprintf("Incorrecct origin %s", or)
		return nil, status.Errorf(codes.Internal, msg)
	}

	out.Volume = nameAndSnap[0]
	out.Snapshot = nameAndSnap[1]
	return &out, nil
}

// GetControllerPlugin get plugin information
func GetControllerPlugin(cfg *ControllerCfg, l *logrus.Entry) (
	cp *ControllerPlugin,
	err error) {
	cp = &ControllerPlugin{}
	lFields := logrus.Fields{
		"node":   "Controller",
		"plugin": "Controller",
	}

	cp.l = l.WithFields(lFields)

	if len(cfg.Iqn) == 0 {
		cfg.Iqn = "iqn.csi.2019-04"
	}
	cp.iqn = cfg.Iqn
	cp.cfg = cfg

	cp.volumesInProcess = make(map[string]bool)

	// Init Storage endpoints
	for _, sConfig := range cfg.StorageEndpoints {
		var storage rest.StorageInterface
		storage, err = rest.NewProvider(&sConfig, l)
		if err != nil {
			cp.l.Warnf("Creating Storage Endpoint failure %+v. Error %s",
				sConfig,
				err)
			continue
		}
		cp.endpoints = append(cp.endpoints, &storage)
		cp.l.Tracef("Add Endpoint %s", sConfig.Name)
	}

	if len(cp.endpoints) == 0 {
		cp.l.Warn("No Endpoints provided in config")
		return nil, errors.New("Unable to create a single endpoint")
	}

	cp.vCap = GetVolumeCapability(supportedVolumeCapabilities)

	// Init tmp volume
	cp.snapReg = cp.cfg.Nodeprefix + "SnapshotRegister"
	_, err = cp.getVolume(cp.snapReg)
	if err == nil {
		return cp, nil
	}
	vd := rest.CreateVolumeDescriptor{
		Name: cp.snapReg,
		Size: minVolumeSize,
	}
	rErr := (*cp.endpoints[0]).CreateVolume(vd)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			cp.l.Warn("Snapshot register already exists.")

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	return cp, nil
}

func (cp *ControllerPlugin) lockVolume(vID string) error {
	var err error
	err = nil
	msg := fmt.Sprintf("Volume is busy", vID)
	errFail := status.Error(codes.Aborted, msg)

	cp.volumesAccess.Lock()
	if cp.volumesInProcess[vID] == false {
		cp.volumesInProcess[vID] = true
	} else {
		err = errFail
	}
	cp.volumesAccess.Unlock()

	return err
}

func (cp *ControllerPlugin) unlockVolume(vID string) error {
	var err error
	err = nil
	msg := fmt.Sprintf("Volume is not locked", vID)
	errFail := status.Error(codes.FailedPrecondition, msg)

	cp.volumesAccess.Lock()
	if cp.volumesInProcess[vID] == true {
		delete(cp.volumesInProcess, vID)
	} else {
		err = errFail
	}
	cp.volumesAccess.Unlock()

	return err
}

func (cp *ControllerPlugin) getStandardId(salt string, name string) string {

	l := cp.l.WithFields(logrus.Fields{
		"func": "getStandardId",
	})

	// Get universal volume ID
	preID := []byte(salt + name)
	rawID := sha256.Sum256(preID)
	id := strings.ToLower(fmt.Sprintf("%X", rawID))
	l.Trace("For %s id is %s", name, id)
	return id
}

func (cp *ControllerPlugin) getRandomName(l int) (s string) {
	var v int64
	out := make([]byte, l)
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567"

	for i := 0; i < l; i++ {
		err := binary.Read(rand.Reader, binary.BigEndian, &v)
		if err != nil {
			cp.l.Fatal(err)
		}
		out[i] = chars[v&31]
	}
	return string(out[:len(out)])
}

func (cp *ControllerPlugin) getRandomPassword(l int) (s string) {
	var v int64
	out := make([]byte, l)
	const chars = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@"

	for i := 0; i < l; i++ {
		err := binary.Read(rand.Reader, binary.BigEndian, &v)
		if err != nil {
			cp.l.Fatal(err)
		}
		out[i] = chars[v&63]
	}
	return string(out[:len(out)])
}

func (cp *ControllerPlugin) getVolume(vID string) (*rest.Volume, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "getVolume",
	})

	l.Tracef("Get volume with id: %s", vID)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if len(vID) == 0 {
		msg := "Volume name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	v, rErr := (*cp.endpoints[0]).GetVolume(vID) // v for Volume

	if rErr != nil {
		switch rErr.GetCode() {
		case rest.RestRequestMalfunction:
			// TODO: correctly process error messages
			err = status.Error(codes.NotFound, rErr.Error())

		case rest.RestRPM:
			err = status.Error(codes.Internal, rErr.Error())
		case rest.RestResourceDNE:
			err = status.Error(codes.NotFound, rErr.Error())
		default:
			err = status.Errorf(codes.Internal, rErr.Error())
		}
		return nil, err
	}
	return v, nil

}

func (cp *ControllerPlugin) createVolumeFromSnapshot(sname string, nvname string) error {
	l := cp.l.WithFields(logrus.Fields{
		"func": "createVolumeFromSnapshot",
	})

	snameT := strings.Split(sname, "_")
	var vname string
	if len(snameT) == 2 {
		vname = snameT[0]
	} else if len(snameT) == 3 {
		vname = snameT[1]
	} else {
		msg := "Unable to obtain volume name from snapshot name"
		l.Warn(msg)
		return status.Error(codes.NotFound, msg)
	}

	rErr := (*cp.endpoints[0]).CreateClone(vname, sname, nvname)
	var err error
	if rErr != nil {
		switch rErr.GetCode() {
		case rest.RestRequestMalfunction:
			// TODO: correctly process error messages
			err = status.Error(codes.NotFound, rErr.Error())
			//return nil, status.Error(codes.Internal, rErr.Error())
		case rest.RestObjectExists:
			err = status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestRPM:
			err = status.Error(codes.Internal, rErr.Error())
		case rest.RestResourceDNE:
			err = status.Error(codes.NotFound, rErr.Error())
		default:
			err = status.Errorf(codes.Internal, rErr.Error())
		}
		return err
	}

	return nil

}

func (cp *ControllerPlugin) createVolumeFromVolume(srcVol string, newVol string) error {
	l := cp.l.WithFields(logrus.Fields{
		"func": "createVolumeFromVolume",
	})
	msg := fmt.Sprintf("Create %s From %s", newVol, srcVol)
	l.Tracef(msg)

	csname, err := cp.createConcealedSnapshot(srcVol)
	if err != nil {
		return err
	}
	err = cp.createVolumeFromSnapshot(*csname, newVol)
	return err
}

// getVolumeSize return size of a volume
func (cp *ControllerPlugin) getVolumeSize(vname string) (int64, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "getVolumeSize",
	})

	v, err := cp.getVolume(vname)

	if err != nil {

		msg := fmt.Sprintf("Internal error %s", err.Error())
		l.Warn(msg)
		err = status.Errorf(codes.Internal, msg)
		return 0, err
	}
	var vSize int64
	vSize, err = strconv.ParseInt((*v).Volsize, 10, 64)
	if err != nil {

		msg := fmt.Sprintf("Internal error %s", err.Error())
		l.Warn(msg)
		err = status.Errorf(codes.Internal, msg)
		return 0, err
	}

	return vSize, nil

}

// CreateVolume create volume with properties
func (cp *ControllerPlugin) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "CreateVolume",
	})

	var err error
	out := csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
		},
	}

	sourceSnapshot := ""
	sourceVolume := ""
	//////////////////////////////////////////////////////////////////////////////
	/// Checks
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}
	vName := req.GetName()

	if len(vName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}

	//TODO: process volume capabilities
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	volumeSize := req.GetCapacityRange().GetRequiredBytes()

	if volumeSize < minVolumeSize {
		maxVSize := req.GetCapacityRange().GetLimitBytes()
		l.Tracef("Minimal volume size %d too small, using max val: %d", volumeSize, maxVSize)
		volumeSize = maxVSize
	}

	if volumeSize < minVolumeSize {
		msg := fmt.Sprintf("Setting volume size to default: %d", volumeSize)
		cp.l.Warn(msg)

		volumeSize = minVolumeSize
	}

	l.Tracef("Create volume %+v of size %+v",
		vName,
		volumeSize)

	//////////////////////////////////////////////////////////////////////////////
	// Check if volume exists

	volumeID := cp.getStandardId(cp.cfg.Salt, vName)

	v, err := cp.getVolume(volumeID)

	if err != nil {

		if codes.NotFound != grpc.Code(err) {
			msg := fmt.Sprintf("Internal error %s", err.Error())
			err = status.Errorf(codes.Internal, msg)
			return nil, err
		}
	}

	// Get info about volume source
	vSource := req.GetVolumeContentSource()
	if vSource != nil {
		out.Volume.ContentSource = req.GetVolumeContentSource()

		if srcSnapshot := vSource.GetSnapshot(); srcSnapshot != nil {
			// Snapshot
			sourceSnapshot = srcSnapshot.GetSnapshotId()
			// Check if snapshot exists
			if _, err = cp.getSnapshot(sourceSnapshot); err != nil {
				return nil, err
			}
			l.Tracef("Sopurce snapshot %s exists.", sourceSnapshot)

		} else if srcVolume := vSource.GetVolume(); srcVolume != nil {
			// Volume
			sourceVolume = srcVolume.GetVolumeId()
			// Check if volume exists
			if _, err = cp.getVolume(sourceVolume); err != nil {
				return nil, err
			}
			cp.l.Tracef("Sopurce volume %s exists.", sourceVolume)
		} else {
			return nil, status.Errorf(codes.Unimplemented,
				"Unable to create volume from other sources")
		}
	}

	// TODO: develop support for different max capacity
	// if voluem exists make shure it has same size
	if v != nil {
		var vSize int64
		vSize, err = strconv.ParseInt((*v).Volsize, 10, 64)
		if vSize != volumeSize {
			msg := fmt.Sprintf("Exists volume with size %d, when requsted for %d", vSize, volumeSize)
			cp.l.Warn(msg)
			err = status.Error(codes.AlreadyExists, msg)
			return nil, err
		}
		// Volume exists
		l.Tracef("Request for the same volume %s with size %d ", volumeID, vSize)

		out.Volume.VolumeId = volumeID
		out.Volume.CapacityBytes = volumeSize

		return &out, nil

	}
	//////////////////////////////////////////////////////////////////////////////
	l.Tracef("req: %+v ", req)

	// Create volume

	vd := rest.CreateVolumeDescriptor{
		Name: volumeID,
		Size: volumeSize,
	}
	var rErr rest.RestError

	if len(sourceSnapshot) > 0 {
		// from snapshot
		err = cp.createVolumeFromSnapshot(sourceSnapshot, volumeID)
		if err != nil {
			return nil, err
		}
		vSize, err := cp.getVolumeSize(volumeID)
		if err != nil {
			return nil, err
		}

		out.Volume.VolumeId = volumeID
		out.Volume.CapacityBytes = vSize

		return &out, nil

	} else if len(sourceVolume) > 0 {
		// from volume
		err = cp.createVolumeFromVolume(sourceVolume, volumeID)
		if err != nil {
			return nil, err
		}
		vSize, err := cp.getVolumeSize(volumeID)
		if err != nil {
			return nil, err
		}

		out.Volume.VolumeId = volumeID
		out.Volume.CapacityBytes = vSize

		return &out, nil

	} else {
		rErr = (*cp.endpoints[0]).CreateVolume(vd)
	}
	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			l.Warn("Specified volume already exists.")

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	out.Volume.VolumeId = volumeID
	out.Volume.CapacityBytes = volumeSize

	return &out, nil

}

// getVolumeSnapshots return array of public volume snapshots
func (cp *ControllerPlugin) getVolumeSnapshots(vname string) ([]rest.SnapshotShort, error) {
	filter := func(s string) bool {
		snameT := strings.Split(s, "_")
		if "c_" == s[:2] {
			return false
		}
		if len(snameT) != 2 {
			return false
		}
		return true
	}
	var snapshots []rest.SnapshotShort

	snapshots, rErr := (*cp.endpoints[0]).ListVolumeSnapshots(
		vname,
		filter)

	if rErr == nil {
		return snapshots, nil
	}

	var err error
	switch code := rErr.GetCode(); code {
	case rest.RestResourceDNE:
		err = status.Error(codes.FailedPrecondition, rErr.Error())
	default:
		err = status.Errorf(codes.Internal, "Unknown internal error")
	}
	return nil, err

}

// getVolumeConcealedSnapshots return array of concealed volume snapshots
func (cp *ControllerPlugin) getVolumeConcealedSnapshots(vname string) ([]rest.SnapshotShort, error) {
	filter := func(s string) bool {
		if "c_" == s[:2] {
			return true
		}
		return false
	}
	snapshots, rErr := (*cp.endpoints[0]).ListVolumeSnapshots(vname, filter)
	if rErr == nil {
		return snapshots, nil
	}

	var err error
	switch code := rErr.GetCode(); code {
	case rest.RestResourceDNE:
		err = status.Error(codes.FailedPrecondition, rErr.Error())
	default:
		err = status.Errorf(codes.Internal, "Unknown internal error")
	}
	return nil, err
}

// getVolumeAllSnapshots return array of concealed volume snapshots
func (cp *ControllerPlugin) getVolumeAllSnapshots(vname string) ([]rest.SnapshotShort, error) {
	filter := func(s string) bool {
		return true
	}
	snapshots, rErr := (*cp.endpoints[0]).ListVolumeSnapshots(vname, filter)
	if rErr == nil {
		return snapshots, nil
	}

	var err error
	switch code := rErr.GetCode(); code {
	case rest.RestResourceDNE:
		err = status.Error(codes.FailedPrecondition, rErr.Error())
	default:
		err = status.Errorf(codes.Internal, "Internal error %s", rErr.Error())
	}
	return nil, err
}

func (cp *ControllerPlugin) gcVolume(vname string) error {
	if err := cp.lockVolume(vname); err != nil {
		return err
	}

	if vname[:2] != "c_" {
		cp.unlockVolume(vname)
		return nil
	}
	dvol, lErr := cp.getVolume(vname)
	if lErr != nil {
		cp.unlockVolume(vname)
		return lErr
	}

	cSnapshots, err := cp.getVolumeConcealedSnapshots(vname)
	if err != nil {
		cp.unlockVolume(vname)
		return err
	}

	for _, snapshot := range cSnapshots {
		if len(snapshot.Clones) > 0 {
			cp.unlockVolume(vname)
			return nil
		}
	}
	cp.unlockVolume(vname)

	lErr = (*cp.endpoints[0]).DeleteVolume(vname, true)
	if lErr != nil {
		return status.Errorf(codes.Internal, lErr.Error())
	}

	if dvol.IsClone {
		or, err := parseOrigin(dvol.Origin)
		if err != nil {
			return err
		}
		if or.Snapshot[:2] == "c_" {
			// volume is made of concealed snapshot
			rErr := (*cp.endpoints[0]).DeleteSnapshot(or.Volume, or.Snapshot)

			if rErr != nil {
				code := rErr.GetCode()
				switch code {
				case rest.RestResourceDNE:
				default:
					return status.Errorf(codes.Internal, rErr.Error())
				}
			}
			// Try to remove parents if they are concealed
			cp.gcVolume(or.Volume)
		}
	}

	return nil
}

// concealVolume tryes to conceal volume
//
// return FailedPrecondition if volume have public snapshots
// checks if volume have and public clones
// conceal volume if it has public clones
// deletes volume if it has no public clones and call concealVolume on its parrent
func (cp *ControllerPlugin) concealVolume(vID string) error {
	l := cp.l.WithFields(logrus.Fields{
		"func": "concealVolume",
	})
	l.Tracef(" %s", vID)

	csl, err := cp.getVolumeAllSnapshots(vID)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	latestSnapshot := csl[len(csl)-1].Name
	cvID := "c_" + vID // concealed volume ID
	err = cp.createVolumeFromSnapshot(latestSnapshot, "c_"+vID)
	if err != nil {
		return err
	}

	if rErr := (*cp.endpoints[0]).PromoteClone(vID, latestSnapshot, cvID); rErr != nil {
		(*cp.endpoints[0]).DeleteClone(vID, latestSnapshot, cvID, false, false)
		msg := fmt.Sprintf("Unable to substitute %s with %s", vID, cvID)
		return status.Error(codes.Internal, msg)
	}

	rErr := (*cp.endpoints[0]).DeleteClone(cvID, latestSnapshot, vID, false, false)

	if rErr != nil {
		eCode := rErr.GetCode()
		switch eCode {
		case rest.RestResourceDNE:
			return nil
		default:
			// Error in process try to recover back
			if rErr := (*cp.endpoints[0]).PromoteClone(cvID, latestSnapshot, vID); rErr != nil {
				(*cp.endpoints[0]).DeleteClone(vID, latestSnapshot, cvID, false, false)

				msg := fmt.Sprintf("Critical ERROR in process of  substitution  %s with %s", vID, cvID)
				l.Error(msg)
				return status.Error(codes.Internal, msg)
			}

		}
		msg := fmt.Sprintf("Unable to substitute %s with %s", vID, cvID)
		return status.Error(codes.Internal, msg)
	}

	return nil
}

// DeleteVolume deletes volume or hides it for later deletion
func (cp *ControllerPlugin) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	l := cp.l.WithFields(logrus.Fields{
		"func": "DeleteVolume",
	})

	var err error
	err = nil
	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		cp.l.Warnf("Unable to delete volume req: %v", req)
		return nil, err
	}

	vID := req.VolumeId
	l.Tracef("Deleting volume %s", vID)

	// Protect volume from modifications
	if err = cp.lockVolume(vID); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	dvol, lErr := cp.getVolume(vID)
	if codes.NotFound == status.Code(lErr) {
		cp.unlockVolume(vID)
		return &csi.DeleteVolumeResponse{}, nil
	}

	if lErr != nil {
		cp.unlockVolume(vID)
		return nil, status.Errorf(codes.Internal, lErr.Error())
	}

	// Try to delete without recursiuon
	lErr = (*cp.endpoints[0]).DeleteVolume(vID, false)

	if lErr == nil {
		//If volume is a clone, GC concealed parents
		cp.unlockVolume(vID)
		if dvol.IsClone {
			or, err := parseOrigin(dvol.Origin)
			if err != nil {
				return nil, err
			}
			if or.Snapshot[:2] == "c_" {
				// volume is made of concealed snapshot
				rErr := (*cp.endpoints[0]).DeleteSnapshot(or.Volume, or.Snapshot)

				if rErr != nil {
					code := rErr.GetCode()
					switch code {
					case rest.RestResourceDNE:
					default:
						cp.unlockVolume(vID)
						return nil, status.Errorf(codes.Internal, rErr.Error())
					}
				}
				// Try to remove parents if they are concealed
				cp.gcVolume(or.Volume)
			}
		}

		return &csi.DeleteVolumeResponse{}, nil
	}

	switch grpc.Code(lErr) {
	case rest.RestResourceDNE:
		cp.unlockVolume(vID)
		return &csi.DeleteVolumeResponse{}, nil
	case rest.RestFailureUnknown:
		cp.unlockVolume(vID)
		return nil, status.Errorf(codes.Internal, lErr.Error())
	case rest.RestResourceBusy:
		// It is likly that volume has snapshots, do nothing
	default:
		cp.unlockVolume(vID)
		return nil, status.Errorf(codes.Internal, "Unknown internal error")
	}

	// check if volume has public snapshots
	psl, lErr := cp.getVolumeSnapshots(vID)
	if lErr != nil {
		switch status.Code(lErr) {
		case codes.FailedPrecondition:
			cp.unlockVolume(vID)
			return &csi.DeleteVolumeResponse{}, nil
		default:
			cp.unlockVolume(vID)
			return nil, status.Errorf(codes.Internal, lErr.Error())
		}
	}
	if len(psl) > 0 {
		// Volume has snapshots, list snapshots in error message
		msg := fmt.Sprintf("Volume %s has snapshots: ", vID)
		for _, s := range psl {
			msg += s.Name + " "
		}
		cp.unlockVolume(vID)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	lErr = cp.concealVolume(vID)
	if lErr != nil {
		return nil, lErr
	}
	return &csi.DeleteVolumeResponse{}, nil

}

// ListVolumes return the list of volumes
func (cp *ControllerPlugin) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	maxEnt := int64(req.GetMaxEntries())
	sToken := req.GetStartingToken()
	////////////////////////////////////////////////////////////////////////////////////////
	// Verify arguments

	if maxEnt < 0 {
		return nil, status.Errorf(codes.Internal, "Number of Entries must not ne negative.")

	}

	if len(sToken) > 0 {
		_, err := cp.getVolume(sToken)

		if err != nil {
			return nil, status.Errorf(codes.Aborted, "%s", err.Error())
		}
	}

	//////////////////////////////////////////////////////////////////////////////

	volumes, err := (*cp.endpoints[0]).ListVolumes()

	if err != nil {
		switch err.GetCode() {
		case rest.RestUnableToConnect:
			return nil, status.Errorf(codes.Internal, "Unable to connect.")

		default:
			return nil, status.Errorf(codes.Internal, "Unable to connect.")
		}
	}

	// Just return all
	if maxEnt == 0 {
		entries := make([]*csi.ListVolumesResponse_Entry, len(volumes))
		for i, name := range volumes {
			entries[i] = &csi.ListVolumesResponse_Entry{
				Volume: &csi.Volume{VolumeId: name},
			}
		}

		return &csi.ListVolumesResponse{
			Entries: entries,
		}, nil
	}

	var iToken int64
	if len(sToken) != 0 {
		iToken, _ = strconv.ParseInt(sToken, 10, 64)
		if int64(len(volumes)) < iToken {
			iToken = 0
		}
	}

	var nextToken = ""

	if int64(len(volumes)) > iToken+maxEnt {
		nextToken = strconv.FormatInt(iToken+maxEnt, 10)
		volumes = volumes[iToken : iToken+maxEnt]

	} else if iToken+maxEnt > int64(len(volumes)) {
		volumes = volumes[iToken:]
	}

	entries := make([]*csi.ListVolumesResponse_Entry, len(volumes))

	for i, name := range volumes {

		entries[i] = &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{VolumeId: name},
		}
	}

	return &csi.ListVolumesResponse{
		Entries:   entries,
		NextToken: nextToken,
	}, nil

}

func (cp *ControllerPlugin) putSnapshotRecord(sID string) error {
	rErr := (*cp.endpoints[0]).CreateSnapshot(cp.snapReg, sID)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err := status.Errorf(codes.Internal, rErr.Error())
			return err

		case rest.RestObjectExists:
			cp.l.Warn("Specified snapshot record already exists.")
			return nil

		default:
			err := status.Errorf(codes.Internal, "Unknown internal error")
			return err
		}
	}
	return nil
}

func (cp *ControllerPlugin) getSnapshotRecordExists(sID string) bool {
	_, rErr := (*cp.endpoints[0]).GetSnapshot(cp.snapReg, sID)

	if rErr != nil {
		return false
		cp.l.Infof("Snapshot record %s DNE", sID)
	}
	cp.l.Infof("Specified snapshot %s exists.", sID)
	return true
}

func (cp *ControllerPlugin) delSnapshotRecord(sID string) error {

	rErr := (*cp.endpoints[0]).DeleteSnapshot(cp.snapReg, sID)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err := status.Errorf(codes.Internal, rErr.Error())
			return err
		case rest.RestObjectExists:
			err := status.Errorf(codes.AlreadyExists, rErr.Error())
			return err
		case rest.RestResourceDNE:
			return nil
		default:
			err := status.Errorf(codes.Internal, "Unknown internal error")
			return err
		}
	}

	return nil
}

// getSnapshot return snapshot datastructure
func (cp *ControllerPlugin) getSnapshot(sID string) (*rest.Snapshot, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "getSnapshot",
	})

	l.Tracef("Get snapshot with id: %s", sID)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if len(sID) == 0 {
		msg := "Snapshot name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	snameT := strings.Split(sID, "_")

	if len(snameT) != 2 {
		msg := "Unable to obtain volume name from snapshot name"
		l.Warn(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	s, rErr := (*cp.endpoints[0]).GetSnapshot(snameT[0], sID)

	if rErr != nil {
		switch rErr.GetCode() {
		case rest.RestRequestMalfunction:
			// TODO: correctly process error messages
			return nil, status.Error(codes.NotFound, rErr.Error())

		case rest.RestRPM:
			return nil, status.Error(codes.Internal, rErr.Error())
		case rest.RestResourceDNE:
			return nil, status.Error(codes.NotFound, rErr.Error())
		default:
			err = status.Errorf(codes.Internal, rErr.Error())
		}
		return nil, err
	}
	return s, nil
}

// createConcealedSnapshot create intermediate snapshot for volume cloning
func (cp *ControllerPlugin) createConcealedSnapshot(vname string) (*string, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "createConcealedSnapshot",
	})

	var sname string

	for i := 0; true; i++ {
		sID := cp.getStandardId("", cp.getRandomName(32))
		sname = fmt.Sprintf("c_%s_%s", vname, sID)

		if _, err := cp.getSnapshot(sname); status.Code(err) == codes.NotFound {
			l.Warn(err.Error())
			break
		}
		if i > 2 {
			return nil, status.Error(codes.Internal, "Unable to pick tmp snapshot name")
		}
	}

	l.Tracef("Snapshot %s", sname)

	rErr := (*cp.endpoints[0]).CreateSnapshot(vname, sname)
	if rErr != nil {
		(*cp.endpoints[0]).DeleteSnapshot(vname, sname)

		return nil, status.Error(codes.Internal, "Unable to create intermidiate snapshot")
	}

	return &sname, nil
}

// CreateSnapshot creates snapshot
func (cp *ControllerPlugin) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "CreateSnapshot",
	})

	l.Trace("Create Snapshot")
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}

	vname := req.GetSourceVolumeId()
	if len(vname) == 0 {
		msg := "Volume name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	sNameRaw := req.GetName()
	// Get universal volume ID

	if len(sNameRaw) == 0 {
		msg := "Snapshot name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	sID := cp.getStandardId(cp.cfg.Salt, sNameRaw)

	sname := fmt.Sprintf("%s_%s", vname, sID)

	bExists := cp.getSnapshotRecordExists(sID)

	if bExists == true {
		cp.l.Debugf("Snapshot record exists!")
		var lerr error
		if _, lerr = cp.getSnapshot(sname); codes.NotFound == grpc.Code(lerr) {
			return nil, status.Error(codes.AlreadyExists, "Exists.")
		}
		if lerr != nil {
			cp.l.Debugf("Err value of checking related property! %s", lerr.Error())
		}
	}

	// Check if volume exists
	//TODO: implement check if snapshot exists
	l.Debugf("Req: %+v ", req)

	// Get size of volume
	var v *rest.Volume
	v, err = cp.getVolume(vname)

	if err != nil {
		return nil, err
	}

	var vSize int64
	vSize, err = strconv.ParseInt((*v).Volsize, 10, 64)

	if err != nil {
		err = status.Errorf(codes.Internal, "Unable to extract volume size.")

	}

	rErr := (*cp.endpoints[0]).CreateSnapshot(vname, sname)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			cp.l.Warn("Specified snapshot already exists.")

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}
	//Make record of created snapshot
	cp.putSnapshotRecord(sID)

	var s *rest.Snapshot // s for snapshot
	s, rErr = (*cp.endpoints[0]).GetSnapshot(vname, sname)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			err = status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
		}
	}

	//Snapshot created successfully
	if rErr == nil {
		layout := "2006-1-2 15:4:5"
		t, err := time.Parse(layout, s.Creation)
		if err != nil {
			msg := fmt.Sprintf("Unable to get snapshot creation time: %s", err)
			cp.l.Warn(msg)
			return nil, status.Errorf(codes.Internal, msg)
		}
		creationTime := &timestamp.Timestamp{

			Seconds: t.Unix(),
		}

		rsp := csi.CreateSnapshotResponse{
			Snapshot: &csi.Snapshot{
				SnapshotId:     sname,
				SourceVolumeId: vname,
				CreationTime:   creationTime,
				ReadyToUse:     true,
				SizeBytes:      vSize,
			},
		}
		cp.l.Tracef("List snapshot resp %+v", rsp)
		return &rsp, nil

	}

	return nil, err
}

// DeleteSnapshot deletes snapshot
func (cp *ControllerPlugin) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	// Check arguments
	l := cp.l.WithFields(logrus.Fields{
		"func": "DeleteSnapshot",
	})

	l.Tracef("Delete Snapshot req: %+v", req)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}

	sname := req.GetSnapshotId()
	if len(sname) == 0 {
		msg := "Snapshot id missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	snameT := strings.Split(sname, "_")

	if len(snameT) != 2 {
		msg := "Unable to obtain volume name from snapshot name"
		l.Warn(msg)
		return &csi.DeleteSnapshotResponse{}, nil
		// TODO: inspect this, according to csi-test
		//return nil, status.Error(codes.InvalidArgument, msg)
	}

	vname := snameT[0]

	//////////////////////////////////////////////////////////////////////////////

	snap, err := cp.getSnapshot(sname)

	if err != nil {

		if codes.NotFound == grpc.Code(err) {
			msg := fmt.Sprintf("Snapshot already deleted %s", sname)

			l.Trace(msg)
			return &csi.DeleteSnapshotResponse{}, nil
		}
	}

	if len(snap.Clones) > 0 {
		msg := fmt.Sprintf("Snapshot %s is a parent of %s", sname, snap.Clones)
		return nil, status.Error(codes.FailedPrecondition, msg)

	}

	rErr := (*cp.endpoints[0]).DeleteSnapshot(vname, sname)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			err = status.Errorf(codes.AlreadyExists, rErr.Error())
			return nil, err

		case rest.RestResourceDNE:

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	// Clean snapshot record
	cp.delSnapshotRecord(snameT[1])

	return &csi.DeleteSnapshotResponse{}, nil

}

// ListSnapshots return the list of valid snapshots
func (cp *ControllerPlugin) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "ListSnapshots",
	})
	msg := fmt.Sprintf("List snapshots %+v", req)
	l.Tracef(msg)
	var err error

	maxEnt := int64(req.GetMaxEntries())

	////////////////////////////////////////////////////////////////////////////////////////
	// Verify arguments

	if maxEnt < 0 {
		return nil, status.Errorf(codes.Internal, "Number of Entries must not be negative.")

	}
	sToken := req.GetStartingToken()

	sname := req.GetSnapshotId()

	if len(sname) != 0 {
		s, err := cp.getSnapshot(sname)

		if err != nil {
			return &csi.ListSnapshotsResponse{
				Entries: []*csi.ListSnapshotsResponse_Entry{},
			}, nil

		}

		snameT := strings.Split(sname, "_")

		iTime, rErr := rest.GetTimeStamp(s.Creation)
		if rErr != nil {
			status.Errorf(codes.Internal, "%s", rErr.Error())
		}
		timeStamp := timestamp.Timestamp{

			Seconds: iTime,
		}

		return &csi.ListSnapshotsResponse{
			Entries: []*csi.ListSnapshotsResponse_Entry{
				{
					Snapshot: &csi.Snapshot{SnapshotId: sname,
						SourceVolumeId: snameT[0],
						CreationTime:   &timeStamp},
				},
			},
		}, nil

	}

	vname := req.GetSourceVolumeId()

	if len(vname) != 0 {
		_, err = cp.getVolume(vname)
		if err != nil {

			if codes.NotFound == grpc.Code(err) {
				msg := fmt.Sprintf("Unable to find volume %s, Err%s", vname, err.Error())
				cp.l.Warn(msg)

				return &csi.ListSnapshotsResponse{
					Entries: []*csi.ListSnapshotsResponse_Entry{},
				}, nil
				return nil, err
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

	}
	l.Trace("Verification done")

	//////////////////////////////////////////////////////////////////////////////
	var rErr rest.RestError

	filter := func(s string) bool {
		snameT := strings.Split(s, "_")
		if len(snameT) != 2 {
			return false
		}
		return true
	}

	var snapshots []rest.SnapshotShort
	if len(vname) == 0 {
		snapshots, rErr = (*cp.endpoints[0]).ListAllSnapshots(filter)
	} else {
		snapshots, rErr = (*cp.endpoints[0]).ListVolumeSnapshots(vname, filter)
	}

	cp.l.Debugf("Obtained snapshots: %d", len(snapshots))
	for i, s := range snapshots {
		cp.l.Debugf("Snap %d, %s", i, s)

	}

	iToken, _ := strconv.ParseInt(sToken, 10, 64)

	if iToken > int64(len(snapshots)) {
		return &csi.ListSnapshotsResponse{
			Entries: []*csi.ListSnapshotsResponse_Entry{},
		}, nil
	}

	//TODO: case with zero snapshots
	if rErr != nil {
		switch rErr.GetCode() {
		case rest.RestUnableToConnect:
			return nil, status.Errorf(codes.Internal, "Unable to connect. Err: %s", rErr.Error())
		default:
			return nil, status.Errorf(codes.Internal, "Unidentified error: %s.", rErr.Error())
		}
	}

	var nextToken = ""

	if maxEnt != 0 || len(sToken) != 0 {
		l.Trace("Listing snapshots of particular parameters")
		if maxEnt == 0 {
			maxEnt = int64(len(snapshots))
		}
		if len(sToken) != 0 {
			iToken, _ = strconv.ParseInt(sToken, 10, 64)
			if int64(len(snapshots)) < iToken {
				iToken = 0
			}
		}

		if int64(len(snapshots)) > iToken+maxEnt {
			nextToken = strconv.FormatInt(iToken+maxEnt, 10)
			snapshots = snapshots[iToken : iToken+maxEnt]

		} else {
			snapshots = snapshots[iToken:]
		}
	}

	entries := make([]*csi.ListSnapshotsResponse_Entry, len(snapshots))

	for i, s := range snapshots {
		cp.l.Debugf("Add snap %s", s.Name)
		timeInt, _ := strconv.ParseInt(s.Properties.Creation, 10, 64)
		timeStamp := timestamp.Timestamp{

			Seconds: timeInt,
		}
		entries[i] = &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SnapshotId:     s.Name,
				SourceVolumeId: s.Volume,
				CreationTime:   &timeStamp,
			},
		}
	}

	return &csi.ListSnapshotsResponse{
		Entries:   entries,
		NextToken: nextToken,
	}, nil

}

// ControllerPublishVolume create iscsi target for the volume
func (cp *ControllerPlugin) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "ControllerPublishVolume",
	})

	l.Tracef("PublishVolume")
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks
	vname := req.GetVolumeId()
	if len(vname) == 0 {
		msg := "Volume id is missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if len(vname) != 64 {
		msg := fmt.Sprintf("Volume id %s is incorrect", vname)
		l.Warn(msg)
		// Get universal volume ID
		vname = cp.getStandardId(cp.cfg.Salt, vname)

	}
	// TODO: verify capabiolity
	caps := req.GetVolumeCapability()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to publish volume req: %v", req)
		return nil, err
	}

	roMode := req.GetReadonly()

	// Check node prefix
	nID := req.GetNodeId()

	if len(nID) == 0 {
		msg := "Node Id must be provided"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if len(cp.cfg.Nodeprefix) > len(nID) {
		msg := "Node Id is too short"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	if strings.HasPrefix(nID, cp.cfg.Nodeprefix) == false {
		msg := "Incorrect Node Id"
		l.Warn(msg)
		return nil, status.Error(codes.NotFound, msg)

	}
	//////////////////////////////////////////////////////////////////////////////

	// Check if volume exists
	_, err = cp.getVolume(vname)

	if err != nil {

		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Create target

	tname := fmt.Sprintf("%s:%s", cp.iqn, vname)

	rErr := (*cp.endpoints[0]).CreateTarget(tname)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			l.Error(rErr.Error())
			err = status.Errorf(codes.AlreadyExists, rErr.Error())
			return nil, err
		case rest.RestResourceDNE:
			msg := fmt.Sprintf("Resource not found: %s", rErr.Error())
			err = status.Errorf(codes.Internal, msg)
			return nil, err

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	// Set Password
	uname := cp.getRandomName(12)
	pass := cp.getRandomPassword(16)
	rErr = (*cp.endpoints[0]).AddUserToTarget(tname, uname, pass)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			err = status.Errorf(codes.AlreadyExists, rErr.Error())
			return nil, err

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	// Attach to target
	var mode string
	if roMode == true {
		mode = "ro"
	} else {
		mode = "wt"
	}

	rErr = (*cp.endpoints[0]).AttachToTarget(tname, vname, mode)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err
		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}
	secrets := map[string]string{"name": uname, "pass": pass}

	secrets["iqn"] = cp.iqn
	secrets["target"] = strings.ToLower(vname)

	var target *rest.Target
	for i := 0; i < 3; i++ {
		target, rErr = (*cp.endpoints[0]).GetTarget(tname)
		if rErr != nil {
			code := rErr.GetCode()
			switch code {
			case rest.RestResourceDNE:
				//According to specification from
				return nil, status.Error(codes.FailedPrecondition, rErr.Error())
			default:
				return nil, status.Errorf(codes.Internal, rErr.Error())
			}
		}
		if target.Active == true {
			l.Tracef("Target %s is active", tname)
			break
		}
		time.Sleep(time.Second)
	}
	if target.Active == false {
		return nil, status.Errorf(codes.Internal, "Unable to make target ready")
	}
	//TODO: add target ip
	// target port
	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: secrets,
	}
	return resp, nil
}

// ControllerUnpublishVolume remove iscsi target for the volume
func (cp *ControllerPlugin) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	l := cp.l.WithFields(logrus.Fields{
		"func": "UnpublishVolume",
	})

	l.Tracef("UnpublishVolume req: %+v", req)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks
	vname := req.GetVolumeId()
	if len(vname) == 0 {
		msg := "Volume name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	tname := fmt.Sprintf("%s:%s", cp.iqn, vname)
	rErr := (*cp.endpoints[0]).DettachFromTarget(tname, vname)

	if rErr != nil {
		c := rErr.GetCode()
		switch c {
		case rest.RestResourceDNE:

		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			status.Errorf(codes.Internal, rErr.Error())
		default:
			status.Errorf(codes.Internal, "Unknown internal error")
		}
	}

	rErr = (*cp.endpoints[0]).DeleteTarget(tname)

	if rErr != nil {
		c := rErr.GetCode()
		switch c {
		case rest.RestResourceDNE:

		case rest.RestResourceBusy:
			//According to specification from
			return nil, status.Error(codes.FailedPrecondition, rErr.Error())
		case rest.RestFailureUnknown:
			err = status.Errorf(codes.Internal, rErr.Error())
			return nil, err

		case rest.RestObjectExists:
			err = status.Errorf(codes.AlreadyExists, rErr.Error())
			return nil, err

		default:
			err = status.Errorf(codes.Internal, "Unknown internal error")
			return nil, err
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks if volume have give capability
func (cp *ControllerPlugin) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	supported := true
	vname := req.GetVolumeId()
	if len(vname) == 0 {
		msg := "Volume name missing in request"
		cp.l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	_, err := cp.getVolume(vname)

	if err != nil {

		return nil, status.Error(codes.NotFound, err.Error())
	}

	vcap := req.GetVolumeCapabilities()

	if vcap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities where not specified")
	}

	for _, c := range vcap {
		m := c.GetAccessMode()
		pass := false
		for _, mode := range supportedVolumeCapabilities {
			if mode == m.Mode {
				pass = true
			}
		}
		if pass == false {
			supported = false
			break
		}

	}

	if supported != true {

	}

	vCtx := req.GetVolumeContext()
	if vcap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume context where not specified")
	}

	resp := &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: cp.vCap,
			VolumeContext:      vCtx,
		},
	}

	return resp, nil

}

// ControllerExpandVolume expands capacity of given volume
func (cp *ControllerPlugin) ControllerExpandVolume(ctx context.Context, in *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity gets storage capacity
func (cp *ControllerPlugin) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	pool, rErr := (*cp.endpoints[0]).GetPool()
	if rErr != nil {
		return nil, status.Error(codes.Internal, rErr.Error())
	}
	var rsp csi.GetCapacityResponse
	var err error
	rsp.AvailableCapacity, err = strconv.ParseInt(pool.Available, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &rsp, nil
}

// capSupported check if capability is supported
func (cp *ControllerPlugin) capSupported(c csi.ControllerServiceCapability_RPC_Type) bool {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		cp.l.Warn("Unknown Capability")
		return false
	}

	for _, cap := range supportedControllerCapabilities {
		if c == cap {
			return true
		}
	}
	cp.l.Debugf("Capability %s isn't supported", c)
	return false
}

// GetVolumeCapability volume related capabilities
func GetVolumeCapability(vcam []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability {
	var out []*csi.VolumeCapability
	for _, c := range vcam {

		vc := csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{Mode: c},
		}

		out = append(out, &vc)
	}

	return out
}

// getControllerServiceCapability incapsulates rpc type of capability to ControllerServiceCapability
func getControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

// ControllerGetCapabilities all capabilities that controller supports
func (cp *ControllerPlugin) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse,
	error,
) {
	cp.l.WithField("func", "ControllerGetCapabilities()").Infof("request: '%+v'", req)

	var capabilities []*csi.ControllerServiceCapability
	for _, c := range supportedControllerCapabilities {
		capabilities = append(capabilities, getControllerServiceCapability(c))
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}
