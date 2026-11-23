/*
Copyright (c) 2019 Open-E, Inc.
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

package rest

import (
	"context"
	"encoding/json"
	"fmt"
	//"reflect"
	//"strconv"
	//"strings"

	log "github.com/sirupsen/logrus"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

func getError(ctx context.Context, body []byte) RestError {
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func": "getError",
	})

	var edata ErrorData
	if err := json.Unmarshal(body, &edata); err != nil {
		bs := string(body)
		msg := fmt.Sprintf("Unable to extract json output from error message: %s", bs)
		l.Warnf("%s", msg)
		return &restError{RestErrorRequestMalfunction, msg}
	} else {
		return ErrorFromErrorT(ctx, &edata.Error, l)
	}
}

func (re *RestEndpoint) unmarshal(resp []byte, ret interface{}) RestError {
	if err := json.Unmarshal(resp, ret); err != nil {
		msg := fmt.Sprintf("Data: %s, Err: %+v.", string(resp[:]), err)
		rErr := GetError(RestErrorRPM, msg)
		// re.l.Warn(rErr.Error())
		return rErr
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Volumes

func (s *RestEndpoint) GetVolume(ctx context.Context, pool string, vname string) (*ResourceVolume, RestError) {
	var resvol ResourceVolume
	rsp := GeneralResponse{Data: &resvol}

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", pool, vname)

	l := s.l.WithFields(log.Fields{
		"func": "GetVolume",
		"url":  addr,
	})

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetVolumeRCode)
	if err != nil {
		msg := fmt.Sprintf("Unable to get volume information")
		l.Warn(msg)
		return nil, GetError(RestErrorRequestMalfunction, msg)
	}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	if stat == CodeOK || stat == CodeNoContent {
		return &resvol, nil
	}

	return nil, getError(ctx, body)
}

func (s *RestEndpoint) CreateVolume(ctx context.Context, pool string, vol CreateVolumeDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/volumes", pool)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "CreateVolume",
		"traceId": ctx.Value("traceId"),
		"url":     addr,
	})

	l.Debugf("sending to pool %s", pool)
	// l.Debugf("Sending data %+v", vol)
	stat, body, err := s.rp.Send(ctx, "POST", addr, vol, CreateVolumeRCode)
	if err != nil {
		s.l.Warnln("Unable to create volume: ", vol.Name)
		return err
	}

	// TODO: we are requesting volume with particular size and JovianDSS returns description of the volume
	// it has created, should we check that one is equal to another?
	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("volume %s creation done", vol.Name)
		return nil
	}

	// TODO: consider case when volume is in process of creation, and not finished yet
	// should we check if it was created successfully

	return getError(ctx, body)
}

// DeleteVolume delete volume, fails if it has snapshots
//
// set rSnapshots to true in order to delete snapshots
func (s *RestEndpoint) DeleteVolume(ctx context.Context, pool string, vname string, data DeleteVolumeDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", pool, vname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "DeleteVolume",
		"url":     addr,
		"section": "rest",
	})

	l.Debugf("Deleting volume %s ", vname)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, data, DeleteVolumeRCode)
	if err != nil {
		s.l.Warnln("Unable to delete volume: ", vname)
		return err
	}

	if stat == CodeOK || stat == CodeNoContent {
		return nil
	}

	return getError(ctx, body)
}

// GetVolumeSnapshot provides information about specific volume snapshot requested
func (s *RestEndpoint) GetVolumeSnapshot(ctx context.Context, pool string, vname string, sname string) (sdp *ResourceSnapshot, err RestError) {
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s", pool, vname, sname)

	l := jcom.LFC(ctx)

	l = s.l.WithFields(log.Fields{
		"section": "rest",
		"func":    "GetVolumeSnapshot",
		"url":     addr,
	})

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetSnapshotRCode)
	if err != nil {
		l.Warnf("Unable to get volume snapshots %+v", err.Error())
		return nil, err
	}

	var snapdata ResourceSnapshot
	rsp := &GeneralResponse{Data: &snapdata}

	if stat == CodeOK || stat == CodeAccepted {
		l.Debug("Obtained volume listing")
		if err = s.unmarshal(body, &rsp); err != nil {
			return nil, err
		}
		return &snapdata, nil
	}

	return nil, getError(ctx, body)
}

// # Create Snapshot from existing volume
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *vid*
//   - *vid* physical volume id as it used by JovianDSS
//   - *desc* data, including snapahot name, that would be transfered to create snapshot
func (s *RestEndpoint) CreateSnapshot(ctx context.Context, pool string, vid string, desc *CreateSnapshotDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots", pool, vid)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
	})

	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CreateSnapshotRCode)
	if err != nil {
		s.l.Warnln("Unable to create snapshot ", desc.SnapshotName)
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("CreateSnapshot %s Done", desc.SnapshotName)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DeleteSnapshot(ctx context.Context, pool string, vname string, sname string, data DeleteSnapshotDescriptor) (err RestError) {
	l := jcom.LFC(ctx)

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s", pool, vname, sname)

	l = l.WithFields(log.Fields{
		"func": "DeleteSnapshot",
		"addr": addr,
	})

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, data, DeleteSnapshotRCode)
	if err != nil {
		s.l.Warnf("Unable to send delete snapshot %s request", sname)
		return err
	}

	if stat == CodeNoContent {
		l.Debugf("Snapshot %s deletion Done", sname)
		return nil
	}

	return getError(ctx, body)
}

// func (s *RestEndpoint) DeleteClone(
//	vname string,
//	sname string,
//	cname string,
//	rChildren bool,
//	rDependent bool) RestError {

// l := s.l.WithFields(logrus.Fields{
// 	"func": "Delete Clone of the Volume",
// })

// data := DeleteClone{
// 	RecursivelyChildren:   rChildren,
// 	RecursivelyDependents: rDependent,
// 	ForceUmount:           false,
// }
// addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones/%s", s.pool, vname, sname, cname)
// msg := fmt.Sprintf("Deleting clone of snapshot %s  volume: %s", sname, vname)
// l.Trace(msg)
// stat, body, err := s.rp.Send("DELETE", addr, data, CreateCloneRCode)

// if err != nil {
// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
// 	l.Warn(msg)
// 	return GetError(RestRequestMalfunction, msg)
// }

// if stat == DeleteCloneRCode {
// 	return nil
// }

// errData, er := s.getError(body)

// if er != nil {
// 	msg := fmt.Sprintf("Unable to extract err message %+v", er)
// 	s.l.Warn(msg)
// 	return GetError(RestRequestMalfunction, msg)
// }

// switch (*errData).Errno {

// case 1:
// 	msg := fmt.Sprintf("Clone %s doesn't exist", cname)
// 	s.l.Warn(msg)
// 	return GetError(RestResourceDNE, msg)
// case 1000:
// 	msg := fmt.Sprintf("Clone %s may have snapshots", cname)
// 	s.l.Warn(msg)
// 	return GetError(RestResourceBusy, msg)
// default:
// 	msg := fmt.Sprintf("Unknown error %d, %s",
// 		(*errData).Errno,
// 		(*errData).Message)
// 	s.l.Warn(msg)
// 	return GetError(RestStorageFailureUnknown, msg)

// }

//	return nil
//}
