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

package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"

	//"encoding/json"
	"os"

	//"errors"
	"fmt"
	//"strconv"
	"strings"
	"sync"

	//"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/timestamp"

	//"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	// "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	jcom "joviandss-kubernetescsi/pkg/common"
	// "joviandss-kubernetescsi/pkg/driver"
	jdrvr "joviandss-kubernetescsi/pkg/driver"
	jrest "joviandss-kubernetescsi/pkg/rest"
	// jtypes "joviandss-kubernetescsi/pkg/types"
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
	minSupportedVolumeSize = 16 * mib
)

var supportedControllerCapabilities = []csi.ControllerServiceCapability_RPC_Type{
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
	csi.ControllerServiceCapability_RPC_GET_CAPACITY,

	// TODO:
	// csi.ControllerServiceCapability_RPC_PUBLISH_READONLY,
}

var supportedVolumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
	// VolumeCapability_AccessMode_UNKNOWN,
	csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
	//csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
	// VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	// VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,

}

// ControllerPlugin provides CSI controller plugin interface
type ControllerPlugin struct {
	l			*log.Logger
	le			*log.Entry
	cfg			*ControllerCfg
	iqn			string
	snapReg			string
	volumesAccess		sync.Mutex
	volumesInProcess	map[string]bool
	
	pool			string
	d			*jdrvr.CSIDriver
	re			jrest.RestEndpoint
	iscsiEendpointCfg	jcom.ISCSIEndpointCfg
	// TODO: add iscsi endpoint
	//iscsiEndpoint    []*rest.StorageInterface
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
func GetControllerPlugin(cp * ControllerPlugin, cfg *jcom.JovianDSSCfg, l *log.Logger) (
	err error,
) {
	os.Exit(1)
	// lFields := logrus.Fields{
	// 	"node":   "Controller",
	// 	"plugin": "Controller",
	// }

	//cp.l = l.WithFields(lFields)

	if len(cfg.ISCSIEndpointCfg.Iqn) == 0 {
		cfg.ISCSIEndpointCfg.Iqn = "iqn.csi.2019-04"
	}
	// cp.iqn = cfg.ISCSIEndpointCfg.Iqn
	// cp.iscsiEendpoint = cfg.ISCSIEndpointCfg

	// cp.volumesInProcess = make(map[string]bool)

	// // Init Storage endpoints
	// re, err = rest.GetEndpoint(&cfg.RestEndpoint, nil)
	// if err != nil {
	// 	cp.l.Warnf("Creating Storage Endpoint failure %+v. Error %s",
	// 		sConfig,
	// 		err)
	// 	continue
	// }
	// cp.re = append(cp.endpoints, &storage)
	// cp.l.Tracef("Add Endpoint %s", sConfig.Name)
	

	// if len(cp.endpoints) == 0 {
	// 	cp.l.Warn("No Endpoints provided in config")
	// 	return errors.New("Unable to create a single endpoint")
	// }

	// cp.vCap = GetVolumeCapability(supportedVolumeCapabilities)

	// Init tmp volume
	// TODO: rethink snapReg
	// cp.snapReg = "CSI-SnapshotRegister"
	// _, err = cp.getVolume(cp.snapReg)
	// if err == nil {
	// 	return nil
	// }
	// vd := rest.CreateVolumeDescriptor{
	// 	Name: cp.snapReg,
	// 	Size: minVolumeSize,
	// }
	// rErr := (*cp.endpoints[0]).CreateVolume(vd)

	// if rErr != nil {
	// 	code := rErr.GetCode()
	// 	switch code {
	// 	case rest.RestResourceBusy:
	// 		// According to specification from
	// 		return status.Error(codes.FailedPrecondition, rErr.Error())
	// 	case rest.RestFailureUnknown:
	// 		err = status.Errorf(codes.Internal, rErr.Error())
	// 		return err

	// 	case rest.RestObjectExists:
	// 		cp.l.Warn("Snapshot register already exists.")

	// 	default:
	// 		err = status.Errorf(codes.Internal, "Unknown internal error")
	// 		return err
	// 	}
	// }

	return nil
}

// GetControllerPlugin get plugin information
func SetupControllerPlugin(cp *ControllerPlugin, cfg *jcom.JovianDSSCfg) (err error) {
	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}
	var e error
	if cp.l, e = jcom.GetLogger(cfg.LLevel, cfg.LDest); e != nil {
		fmt.Fprintln(os.Stderr, "Unable to init loging because:", e)
		os.Exit(1)
	}
	cp.le =cp.l.WithFields(log.Fields{"section": "controller", "traceId": "setup" })

	if cp.d, err = jdrvr.NewJovianDSSCSIDriver(&cfg.RestEndpointCfg, cp.le); err != nil {
		return err
	}

	jrest.SetupEndpoint(&cp.re, &cfg.RestEndpointCfg, cp.le)


	if len(cfg.ISCSIEndpointCfg.Iqn) == 0 {
		cfg.ISCSIEndpointCfg.Iqn = "iqn.csi.2019-04"
	}
	cp.iqn = cfg.ISCSIEndpointCfg.Iqn
	cp.iscsiEendpointCfg = cfg.ISCSIEndpointCfg
	cp.pool = cfg.Pool
	// cp.volumesInProcess = make(map[string]bool)

	// // Init Storage endpoints
	// re, err = rest.GetEndpoint(&cfg.RestEndpoint, nil)
	// if err != nil {
	// 	cp.l.Warnf("Creating Storage Endpoint failure %+v. Error %s",
	// 		sConfig,
	// 		err)
	// 	continue
	// }
	// cp.re = append(cp.endpoints, &storage)
	// cp.l.Tracef("Add Endpoint %s", sConfig.Name)
	

	// if len(cp.endpoints) == 0 {
	// 	cp.l.Warn("No Endpoints provided in config")
	// 	return errors.New("Unable to create a single endpoint")
	// }

	// cp.vCap = GetVolumeCapability(supportedVolumeCapabilities)

	// Init tmp volume
	// TODO: rethink snapReg
	// cp.snapReg = "CSI-SnapshotRegister"
	// _, err = cp.getVolume(cp.snapReg)
	// if err == nil {
	// 	return nil
	// }
	// vd := rest.CreateVolumeDescriptor{
	// 	Name: cp.snapReg,
	// 	Size: minVolumeSize,
	// }
	// rErr := (*cp.endpoints[0]).CreateVolume(vd)

	// if rErr != nil {
	// 	code := rErr.GetCode()
	// 	switch code {
	// 	case rest.RestResourceBusy:
	// 		// According to specification from
	// 		return status.Error(codes.FailedPrecondition, rErr.Error())
	// 	case rest.RestFailureUnknown:
	// 		err = status.Errorf(codes.Internal, rErr.Error())
	// 		return err

	// 	case rest.RestObjectExists:
	// 		cp.l.Warn("Snapshot register already exists.")

	// 	default:
	// 		err = status.Errorf(codes.Internal, "Unknown internal error")
	// 		return err
	// 	}
	// }

	return nil
}

func (cp *ControllerPlugin) lockVolume(vID string) error {
	var err error
	err = nil
	msg := fmt.Sprintf("Volume %s is busy", vID)
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
	msg := fmt.Sprintf("Volume %s is not locked", vID)
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

func (cp *ControllerPlugin) getStandardID(name string) string {
	l := cp.l.WithFields(log.Fields{
		"func": "getStandardID",
	})

	// Get universal volume ID
	preID := []byte(name)
	rawID := sha256.Sum256(preID)
	id := strings.ToLower(fmt.Sprintf("%X", rawID))
	l.Tracef("For %s id is %s", name, id)
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
	return string(out[:])
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
	return string(out[:])
}

func (cp *ControllerPlugin) getVolume(ctx context.Context, vID string) (*jrest.ResourceVolume, error) {
	// return nil, nil
	l := cp.l.WithField("traceId", ctx.Value("traceId"))
		//Value("traceId").(string))

	l.Debugf("context %+v", ctx)
	l.Debugf("Get volume with id: %s", vID)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if len(vID) == 0 {
		msg := "Volume name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	v, rErr := cp.re.GetVolume(ctx, cp.pool, vID) // v for Volume

	l.Debugf("%+v\n",v)
	l.Debugf("%+v\n",rErr)
	l.Debugf("%+v\n",rErr.GetCode())
	if rErr != nil {
		switch rErr.GetCode() {
		case jrest.RestErrorRequestMalfunction:
			// TODO: correctly process error messages
			err = status.Error(codes.NotFound, rErr.Error())
		case jrest.RestErrorRPM:
			err = status.Error(codes.Internal, rErr.Error())
		case jrest.RestErrorResourceDNE:
			err = status.Error(codes.NotFound, rErr.Error())
		default:
			err = status.Error(codes.Internal, rErr.Error())
		}
		return nil, err
	}
	return v, nil
}


func (cp *ControllerPlugin) createVolumeFromSnapshot(sd jdrvr.SnapshotDesc, nvid jdrvr.VolumeDesc) error {
//	l := cp.l.WithFields(logrus.Fields{
//		"func": "createVolumeFromSnapshot",
//	})
//
//	snameT := strings.Split(sname, "_")
//	var vname string
//	if len(snameT) == 2 {
//		vname = snameT[0]
//	} else if len(snameT) == 3 {
//		vname = snameT[1]
//	} else {
//		msg := "Unable to obtain volume name from snapshot name"
//		l.Warn(msg)
//		return status.Error(codes.NotFound, msg)
//	}
//
//	rErr := (*cp.endpoints[0]).CreateClone(vname, sname, nvname)
//	var err error
//	if rErr != nil {
//		switch rErr.GetCode() {
//		case rest.RestRequestMalfunction:
//			// TODO: correctly process error messages
//			err = status.Error(codes.NotFound, rErr.Error())
//			// return nil, status.Error(codes.Internal, rErr.Error())
//		case rest.RestObjectExists:
//			err = status.Error(codes.FailedPrecondition, rErr.Error())
//		case rest.RestRPM:
//			err = status.Error(codes.Internal, rErr.Error())
//		case rest.RestResourceDNE:
//			err = status.Error(codes.NotFound, rErr.Error())
//		default:
//			err = status.Errorf(codes.Internal, rErr.Error())
//		}
//		return err
//	}
//
	return nil
}

func (cp *ControllerPlugin) createVolumeFromVolume(sd jdrvr.SnapshotDesc, nvid jdrvr.VolumeDesc) error {
	l := cp.l.WithFields(log.Fields{
		"func": "createVolumeFromVolume",
	})

	msg := fmt.Sprintf("Create %s From %s", sd.Name(), nvid.Name())
	l.Tracef(msg)

	// csname, err := cp.createConcealedSnapshot(srcVol)
	// if err != nil {
	// 	return err
	// }
	err := cp.createVolumeFromSnapshot(sd, nvid)
	return err
}

// getVolumeSize return size of a volume
func (cp *ControllerPlugin) getVolumeSize(vname string) (int64, error) {
	// l := cp.l.WithFields(logrus.Fields{
	// 	"func": "getVolumeSize",
	// })

	// v, err := cp.getVolume(vname)
	// if err != nil {

	// 	msg := fmt.Sprintf("Internal error %s", err.Error())
	// 	l.Warn(msg)
	// 	err = status.Errorf(codes.Internal, msg)
	// 	return 0, err
	// }
	// var vSize int64
	// vSize, err = strconv.ParseInt((*v).Volsize, 10, 64)
	// if err != nil {

	// 	msg := fmt.Sprintf("Internal error %s", err.Error())
	// 	l.Warn(msg)
	// 	err = status.Errorf(codes.Internal, msg)
	// 	return 0, err
	// }
	
	return 0, nil
}

func (cp *ControllerPlugin) createNewVolume(ctx context.Context, nvd *jdrvr.VolumeDesc, capr *csi.CapacityRange, vSource *csi.VolumeContentSource) (volumeSize int64, csierr error) {

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "createNeVolume",
		"section": "controller",
	})

	var err jrest.RestError = nil
	if vSource != nil {
		l.Debugf("Creating volume from source %+v", vSource)
		if srcSnapshot := vSource.GetSnapshot(); srcSnapshot != nil {
			// Snapshot
			sourceSnapshotID := srcSnapshot.GetSnapshotId()
			sd, err := jdrvr.NewSnapshotDescFromCSIID(sourceSnapshotID)	
			if err == nil {
				l.Debugf("Creating volume %s from snapshot %s", nvd.Name(), sd.Name())
				err = cp.d.CreateVolumeFromSnapshot(ctx, cp.pool, sd, nvd)
			} else {
				return 0, status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to identify snapshot source %s", sourceSnapshotID))
			}

		} else if srcVolume := vSource.GetVolume(); srcVolume != nil {
			// Volume
			sourceVolumeID := srcVolume.GetVolumeId()
			// Check if volume exists
			vd, csierr := jdrvr.NewVolumeDescFromVDS(sourceVolumeID)
			if csierr == nil {
				l.Debugf("Creating volume %s from volume %s", nvd.Name(), vd.Name())
				err = cp.d.CreateVolumeFromVolume(ctx, cp.pool, vd, nvd)
			} else {
				return 0, status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to identify volume source %s", sourceVolumeID))
			}

		} else {
			return 0, status.Errorf(codes.Unimplemented, "Unable to create volume from other sources")
		}
	} else {
		l.Debugf("required bytes %d, limit bytes %d", capr.GetRequiredBytes(), capr.GetLimitBytes())

		if capr.GetLimitBytes() == capr.GetRequiredBytes() {
			volumeSize = capr.GetLimitBytes()
		} else if capr.GetRequiredBytes() < minSupportedVolumeSize {
			volumeSize = minSupportedVolumeSize
		} else {
			volumeSize = capr.GetRequiredBytes()
		}

		err = cp.d.CreateVolume(ctx, cp.pool, nvd, volumeSize)
	}

	switch jrest.ErrCode(err)  {
	case jrest.RestErrorResourceBusy:
		return 0, status.Error(codes.FailedPrecondition, err.Error())
	case jrest.RestErrorResourceDNE:
		return 0, status.Error(codes.NotFound, err.Error())
	case jrest.RestErrorResourceExists:
		l.Warn("Specified volume already exists.")
		return 0, status.Errorf(codes.AlreadyExists, err.Error())
	case jrest.RestErrorOutOfSpace:
		emsg := fmt.Sprintf("Unable to create volume %s, storage out of space", nvd.Name())
		l.Warn(emsg)
		return 0, status.Errorf(codes.ResourceExhausted, emsg)
	case jrest.RestErrorOk:
		l.Debugf("Volume %s created", nvd.Name())
		return volumeSize, nil
	default:
		return 0, status.Errorf(codes.Internal, err.Error())
	}
}

// VolumeComply checks if volume with specified properties exists
//
// if volume with same name exists yet does not fall into requirments it fails with ALLREADY_EXISTS
// if volume does not exists it fails with NOT_FOUND
// if volume do exists and fit requirmnets it will return csi volume struct and nil as error
func (cp *ControllerPlugin) VolumeComply(ctx context.Context, vd *jdrvr.VolumeDesc, caprage *csi.CapacityRange, source *csi.VolumeContentSource) (*int64, error ) {

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "VolumeComply",
		"section": "controller",
	})

	l.Debugf("Checking if volume %s exists and comply with requirments %+v %+v", vd.Name(), caprage, source)

	vdata, jerr := cp.d.GetVolume(ctx, cp.pool, vd);
	
	if jerr != nil {
		if jerr.GetCode() == jrest.RestErrorResourceDNE {
			return nil, status.Errorf(codes.NotFound, jerr.Error())
		}
	}
	
	s := vdata.GetSize()

	if caprage != nil {
		if minSize := caprage.GetRequiredBytes(); minSize > 0 && minSize > vdata.GetSize() {
			return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Existing volume %s have size %d that is less then minimal requested limit %d", vd.Name(), caprage.GetRequiredBytes(), minSize))
		}
		
		if maxSize := caprage.GetLimitBytes(); maxSize > 0 && maxSize < vdata.GetSize() {
			return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Existing volume %s have size %d that is more then upper limit %d", vd.Name(), caprage.GetRequiredBytes(), maxSize))
		}
	}

	if source != nil {
		
		if sv := source.GetVolume(); sv != nil {

			if ov := vdata.OriginVolume(); ov != sv.GetVolumeId() {
				if len(ov) == 0 && len(sv.GetVolumeId()) > 0 {
					return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Volume with name %s exists and it is not derived from any volume", vd.Name()))
				}
				return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Existing volume %s is derived from volume %s that is different from requested volume %s", vd.Name(), vdata.OriginVolume(), sv.GetVolumeId()))
			}
		}

		if sv := source.GetSnapshot(); sv != nil {

			if vol, err := jdrvr.NewVolumeDescFromVDS(vdata.OriginVolume()); err != nil {
				return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Volume %s exists, but driver is not able to identify correctly its origin %s", vd.Name(), vdata.OriginVolume()))
			} else {
				if snap, err := jdrvr.NewSnapshotDescFromSDS(vol, vdata.OriginSnapshot()); err != nil {
					return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Volume %s exists, but driver is not able to identify correctly its origin snapshot %s", vd.Name(), vdata.OriginSnapshot()))
				} else {
					if snap.CSIID() != sv.GetSnapshotId() {
						return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Existing volume %s is derived from snapshot %s, not from requested one %s", vd.Name(), snap.CSIID(), sv.GetSnapshotId()))
					}
				}
			}
		}
	}

	l.Debugf("Volume %s check done, volume comply", vd.Name())

	return &s, status.Errorf(codes.OK, "")
}

// CreateVolume create volume with properties
func (cp *ControllerPlugin) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	
	l := cp.l.WithFields(log.Fields{
		"request": "CreateVolume",
		"func": "CreateVolume",
		"section": "controller",
	})
	ctx = jcom.WithLogger(ctx, l)

	var err error
	out := csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
		},
	}

	///////////////////////////////////////////////////////////////////////
	/// Check capability
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}
	// vName := req.GetName()
	var nvid *jdrvr.VolumeDesc
	if nvid, err = jdrvr.NewVolumeDescFromName(req.GetName()); err != nil {
		return nil, err
	}

	// TODO: process volume capabilities
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	//volumeSize := req.GetCapacityRange().GetRequiredBytes()
	maxVSize := req.GetCapacityRange().GetLimitBytes()

	if maxVSize > 0 {
		if maxVSize < minSupportedVolumeSize {
			return nil, status.Error(codes.OutOfRange, fmt.Sprintf("Volume size must be at least %d bytes", minSupportedVolumeSize))
		}
	}

	l.Debugf("Create volume capability check done")


	// Check if volume exists and comply with requirments 
	vsize, err := cp.VolumeComply(ctx, nvid, req.GetCapacityRange(), req.GetVolumeContentSource())
	switch status.Code(err) {
		case codes.AlreadyExists:
			return nil, err
		case codes.OK:
			l.Debugf("Volume %s already exist and comply with requirmnets", nvid.Name())
			out.Volume.VolumeId = nvid.VDS()
			out.Volume.CapacityBytes = *vsize
			return &out, nil
		case codes.NotFound:
			l.Debugf("Volume %s do not exists, creating", nvid.Name())
		default:
			return nil, status.Error(codes.Unknown, fmt.Sprintf("Unable to identify if volume exists or not: %s", err.Error()))
	}

	if vSize, err := cp.createNewVolume(ctx, nvid, req.GetCapacityRange(), req.GetVolumeContentSource()); err != nil {
		return nil, err
	} else {
		out.Volume.VolumeId = nvid.VDS()
		out.Volume.CapacityBytes = vSize
	}

	return &out, nil
}

// getVolumeSnapshots return array of public volume snapshots
//func (cp *ControllerPlugin) getVolumeSnapshots(vname string) ([]jrest.SnapshotShort, error) {
//	return nil, nil
	// filter := func(s string) bool {
	// 	snameT := strings.Split(s, "_")
	// 	if "c_" == s[:2] {
	// 		return false
	// 	}
	// 	if len(snameT) != 2 {
	// 		return false
	// 	}
	// 	return true
	// }
	// var snapshots []rest.SnapshotShort

	// snapshots, rErr := (*cp.endpoints[0]).ListVolumeSnapshots(
	// 	vname,
	// 	filter)

	// if rErr == nil {
	// 	return snapshots, nil
	// }

	// var err error
	// switch code := rErr.GetCode(); code {
	// case rest.RestResourceDNE:
	// 	err = status.Error(codes.FailedPrecondition, rErr.Error())
	// default:
	// 	err = status.Errorf(codes.Internal, "Unknown internal error")
	// }
	// return nil, err
//}


// DeleteVolume deletes volume or hides it for later deletion
func (cp *ControllerPlugin) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	
	l := cp.l.WithFields(log.Fields{
		"request": "DeleteVolume",
		"func": "DeleteVolume",
		"section": "controller",
	})
	ctx = jcom.WithLogger(ctx, l)

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
		err := status.Errorf(codes.Internal, "Capability is not supported.")
		cp.l.Warnf("Unable to delete volume req: %v", req)
		return nil, err
	}

	if vd, rerr := jdrvr.NewVolumeDescFromVDS(req.VolumeId); rerr == nil {

		l.Debugf("Deleting volume %s", vd.Name())

		// Try to delete without recursiuon
		if err := cp.d.DeleteVolume(ctx, cp.pool, vd); err == nil {
			return &csi.DeleteVolumeResponse{}, nil
		} else {
			switch err.GetCode() {
			case jrest.RestErrorResourceBusy:
				return nil, status.Error(codes.FailedPrecondition, err.Error())
			case jrest.RestErrorResourceDNE:
				l.Warnf("Volume %s was deleted before", vd)
				return &csi.DeleteVolumeResponse{}, nil
			default:
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Volume id %s has wrong name format", req.VolumeId))
	}
}

//ListVolumes return the list of volumes
func (cp *ControllerPlugin) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	var rErr jrest.RestError
	var token *jdrvr.CSIListingToken
	var resp csi.ListVolumesResponse

	l := cp.l.WithFields(log.Fields{
		"request": "ListVolumes",
		"func": "ListVolumes",
		"section": "controller",
	})
	ctx = jcom.WithLogger(ctx, l)
	
	l.Debugf("Request: %+v", req)

	maxEnt := int64(req.GetMaxEntries())
	startingToken := req.GetStartingToken()
	token, rErr = jdrvr.NewCSIListingTokenFromTokenString(startingToken)
	
	if rErr != nil {
		return nil, status.Errorf(codes.Aborted, "Unable to operate with token %s Err: %s", startingToken, rErr.Error())
	}

	if maxEnt < 0 {
		return nil, status.Errorf(codes.Internal, "Number of Entries must not be negative.")
	}

	if  volList, ts, rErr := cp.d.ListAllVolumes(ctx, cp.pool, int(maxEnt), *token); rErr != nil {
		l.Debugf("Unable to comlete listing %s", rErr.Error())
		return nil, status.Errorf(codes.Internal, "Unable to complete listing request: %s", rErr.Error()) 
	} else {
		if ts != nil {
			resp.NextToken = ts.Token()
		}

		if err := completeListResponseFromVolume(ctx, &resp, volList); err != nil {
			return nil, err
		} else {
			return &resp, nil
		}
	}

}

// CreateSnapshot creates snapshot
func (cp *ControllerPlugin) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	
	l := cp.l.WithFields(log.Fields{
		"request": "CreateSnapshot",
		"func": "CreateSnapshot",
	})

	ctx = jcom.WithLogger(ctx, l)

	l.Debug("request: %+s", *req)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}

	//////////////////////////////////////////////////////////////////////////////

	vd, err := jdrvr.NewVolumeDescFromCSIID(req.GetSourceVolumeId())
	if err != nil {
		return nil, err
	}

	sd := jdrvr.NewSnapshotDescFromName(vd, req.GetName())


	rErr := cp.d.CreateSnapshot(ctx, cp.pool, vd, sd)

	switch jrest.ErrCode(rErr)  {
	case jrest.RestErrorResourceBusy:
		return nil, status.Error(codes.Aborted, err.Error())
	case jrest.RestErrorResourceDNE:
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case jrest.RestErrorResourceExists:
		l.Warn("Specified snapshot already exists.")
	case jrest.RestErrorOutOfSpace:
		emsg := fmt.Sprintf("Unable to create snapshot %s for volume %s, storage out of space", sd.Name(), vd.Name())
		l.Warn(emsg)
		return nil, status.Errorf(codes.ResourceExhausted, emsg)
	case jrest.RestErrorOk:
		l.Debugf("Snapshot %s was created", sd.Name())
	default:
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	snap, rErr := cp.d.GetSnapshot(ctx, cp.pool, vd, sd)
	
	switch jrest.ErrCode(rErr)  {
	case jrest.RestErrorResourceBusy:
		return nil, status.Error(codes.Aborted, err.Error())
	case jrest.RestErrorResourceDNE:
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case jrest.RestErrorOk:
		l.Debugf("Got snapshot %s info %+v", sd.Name(), *snap)
	default:
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	creationTime := &timestamp.Timestamp{
		Seconds: snap.Creation.Unix(),
	}
	
	rsp := csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     sd.CSIID(),
			SourceVolumeId: vd.CSIID(),
			CreationTime:   creationTime,
			ReadyToUse:     true,
			SizeBytes:      snap.VolSize,
		},
	}
	return &rsp, nil
}

// DeleteSnapshot deletes snapshot
func (cp *ControllerPlugin) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {

	l := cp.l.WithFields(log.Fields{
		"request": "DeleteSnapshot",
		"func": "DeleteSnapshot",
	})

	ctx = jcom.WithLogger(ctx, l)

	l.Debug("request: %+s", *req)
	var err error
	

	//////////////////////////////////////////////////////////////////////////////
	/// Checks
	if false == cp.capSupported(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
		err = status.Errorf(codes.Internal, "Capability is not supported.")
		l.Warnf("Unable to create volume req: %v", req)
		return nil, err
	}
	//////////////////////////////////////////////////////////////////////////////

	sd, err := jdrvr.NewSnapshotDescFromCSIID(req.GetSnapshotId())
	if err != nil {
		return nil, err
	}

	ld := sd.GetVD()

	rErr := cp.d.DeleteSnapshot(ctx, cp.pool, ld, sd)

	switch jrest.ErrCode(rErr)  {
	case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusySnapshotHasClones:
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case jrest.RestErrorResourceDNE:
		l.Warnf("snapshot %s do not exists", sd.Name())
	case jrest.RestErrorOk:
		l.Debugf("snapshot %s was deleted before", sd.Name())
	default:
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots return the list of valid snapshots
func (cp *ControllerPlugin) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (rsp *csi.ListSnapshotsResponse, err error) {

	var rErr jrest.RestError
	var token *jdrvr.CSIListingToken
	var resp csi.ListSnapshotsResponse

	l := cp.l.WithFields(log.Fields{
		"request": "ListSnapshtos",
		"func": "ListSnapshots",
	})
	ctx = jcom.WithLogger(ctx, l)

	l.Debugf("Request: %+v", req)

	maxEnt := int64(req.GetMaxEntries())
	sourceVolumeId := req.GetSourceVolumeId()
	snapshotId := req.GetSnapshotId()
	startingToken := req.GetStartingToken()
	token, rErr = jdrvr.NewCSIListingTokenFromTokenString(startingToken)

	if rErr != nil {
		return nil, status.Errorf(codes.Aborted, "Unable to operate with token %s Err: %s", startingToken, rErr.Error())
	}

	if maxEnt < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Number of Entries must not be negative.")
	}

	if len(sourceVolumeId) > 0 && len(snapshotId) == 0 {
		l.Debugf("for volume %s", sourceVolumeId)
		if vd, err := jdrvr.NewVolumeDescFromCSIID(sourceVolumeId); err != nil {
			return nil, err
		} else {
			if  snapList, ts, rErr := cp.d.ListVolumeSnapshots(ctx, cp.pool, vd, int(maxEnt), *token); rErr != nil {
				return nil, status.Errorf(codes.Internal, "Unable to complete listing request: %s", rErr.Error()) 
			} else {
				if ts != nil {
					resp.NextToken = ts.Token()
				}
				if err = completeListResponseFromVolumeSnapshot(ctx, &resp, snapList, vd); err != nil {
					return nil, err
				} else {
					return &resp, nil
				}
			}
		}
	} else if len(snapshotId) > 0 {
		if sd, err := jdrvr.NewSnapshotDescFromCSIID(snapshotId); err != nil {
			return nil, err
		} else {
			ld := sd.GetVD()
			if len(sourceVolumeId) > 0 && sourceVolumeId != ld.CSIID() {
				return nil, status.Errorf(codes.FailedPrecondition, "Specified snapshot %s with id %s is not related to volume %s with id %s", sd.Name(), sd.CSIID(), ld.Name(), ld.CSIID())
			}
			if snap, rErr := cp.d.GetSnapshot(ctx, cp.pool, ld, sd); rErr != nil {
				if jrest.ErrCode(rErr) == jrest.RestErrorResourceDNE {
					return nil, status.Error(codes.InvalidArgument, rErr.Error())
				}

				entry := csi.ListSnapshotsResponse_Entry{
					Snapshot: &csi.Snapshot{
						SnapshotId:     sd.CSIID(),
						SourceVolumeId: ld.CSIID(),
						CreationTime:   timestamppb.New(snap.Creation),
						ReadyToUse:	true,
					},
				}
				resp.Entries = append(resp.Entries, &entry)
			}
		}

		l.Debugf("get snapshot %s", snapshotId)
	} else {
		l.Debugln("listing all snapshots")
		if  snapList, ts, rErr := cp.d.ListAllSnapshots(ctx, cp.pool, int(maxEnt), *token); err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to complete listing request: %s", rErr.Error()) 
		} else {
			if ts != nil {
				resp.NextToken = ts.Token()
			}
			if err = completeListResponseFromSnapshotShort(ctx, &resp, snapList); err != nil {
				return nil, err
			} else {
				return &resp, nil
			}
		}
	}
	return nil, nil
}

// ControllerPublishVolume create iscsi target for the volume
func (cp *ControllerPlugin) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	
	l := cp.l.WithFields(log.Fields{
		"request": "ControllerPublishVolume",
		"func": "ControllerPublishVolume",
	})
	ctx = jcom.WithLogger(ctx, l)

	l.Debugf("Publish volume request %+v", req)
	var err error

	vd, err := jdrvr.NewVolumeDescFromCSIID(req.GetVolumeId())
	if err != nil {
		return nil, err
	}
	roMode := req.GetReadonly()

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

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

	//////////////////////////////////////////////////////////////////////////////

	iscsiContext, rErr := cp.d.PublishVolume(ctx, cp.pool, vd, cp.iqn, roMode)

	switch jrest.ErrCode(rErr) {
	case jrest.RestErrorOk:
		resp := csi.ControllerPublishVolumeResponse{
			PublishContext: *iscsiContext,
		}
		return &resp, nil
	case jrest.RestErrorResourceBusy:
		// According to specification from
		return nil, status.Error(codes.FailedPrecondition, rErr.Error())
	case jrest.RestErrorFailureUnknown:
		err = status.Errorf(codes.Internal, rErr.Error())
		return nil, err
	case jrest.RestErrorResourceExists:
		// TODO: handle	FAILED_PRECONDITION
		// Indicates that a volume corresponding to the specified volume_id has already been published at another node and does not have MULTI_NODE volume capability.
		// If this error code is returned, the Plugin SHOULD specify the node_id of the node at which the volume is published as part of the gRPC status.message.
		// TODO: handle ALREADY_EXISTS
		// Indicates that a volume corresponding to the specified volume_id has already been published at the node corresponding to the specified node_id but is 
		// incompatible with the specified volume_capability or readonly flag .
		err = status.Errorf(codes.AlreadyExists, rErr.Error())
		return nil, err
	case jrest.RestErrorResourceDNE:
		msg := fmt.Sprintf("Resource not found: %s", rErr.Error())
		err = status.Errorf(codes.NotFound, msg)
		return nil, err

	default:
		err = status.Errorf(codes.Internal, "Unknown internal error")
		return nil, err
	}
}

// ControllerUnpublishVolume remove iscsi target for the volume
func (cp *ControllerPlugin) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	l := cp.l.WithFields(log.Fields{
		"request": "ControllerUnpublishVolume",
		"func": "ControllerUnpublishVolume",
	})
	ctx = jcom.WithLogger(ctx, l)

	l.Debugf("UnpublishVolume req: %+v", req)

	if vd, err := jdrvr.NewVolumeDescFromCSIID(req.GetVolumeId()); err != nil {
		return nil, err
	} else {
		rErr := cp.d.UnpublishVolume(ctx, cp.pool, cp.iqn, vd); 
		switch jrest.ErrCode(rErr) {
		case jrest.RestErrorOk, jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNETarget:
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		default:
			return nil, status.Errorf(codes.Internal, "Unable to unpublish volume %s because of %s", vd.Name(), rErr.Error())
		}
	}
}

// ValidateVolumeCapabilities checks if volume have give capability
func (cp *ControllerPlugin) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	l := cp.l.WithFields(log.Fields{
		"request": "ValidateVolumeCapabilities",
		"func": "ValidateVolumeCapabilitiese",
	})

	ctx = jcom.WithLogger(ctx, l)
	supported := true

	vd, err := jdrvr.NewVolumeDescFromCSIID(req.GetVolumeId())
	if err != nil {
		return nil, err
	}

	_, rErr := cp.d.GetVolume(ctx, cp.pool, vd)
	switch jrest.ErrCode(rErr) {
	case jrest.RestErrorOk:
		l.Debugf("volume %s present", vd.Name())
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEVolume:
		return nil, status.Error(codes.NotFound, rErr.Error())
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to verify volume %s because of %s ", vd.Name(), rErr.Error()))
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

// ControllerGetVolume provides current information about the volume
func (cp *ControllerPlugin) ControllerGetVolume(ctx context.Context, in *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity gets storage capacity
func (cp *ControllerPlugin) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	// TODO: add capability check
	pool, rErr := cp.d.GetPool(ctx, cp.pool)
	if rErr != nil {
		return nil, status.Error(codes.Internal, rErr.Error())
	}
	var rsp csi.GetCapacityResponse

	rsp.AvailableCapacity = pool.Available
	rsp.MinimumVolumeSize = wrapperspb.Int64(minSupportedVolumeSize)
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
