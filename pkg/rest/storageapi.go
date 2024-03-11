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
	"time"

	log "github.com/sirupsen/logrus"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	jcom "joviandss-kubernetescsi/pkg/common"
)


type SnapshotDescriptor struct {
	VName   string
	SName   string
	Created string
}

func getError(ctx context.Context, body []byte) RestError {
	
	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func": "getError",
	})
	
	var edata ErrorData
	if err := json.Unmarshal(body, &edata); err != nil {
		bs := fmt.Sprintf(string(body[:len(body)]))
		msg := fmt.Sprintf("Unable to extract json output from error message: %s", bs)
		l.Warnf(msg)
		return &restError{RestErrorRequestMalfunction, msg}
	} else {

		return ErrorFromErrorT(ctx, &edata.Error, l)
	}
}

func (re *RestEndpoint) unmarshal(resp []byte, ret interface{}) (RestError)  {
	if err := json.Unmarshal(resp, ret); err != nil {
		msg := fmt.Sprintf("Data: %s, Err: %+v.", string(resp[:len(resp)]), err)
		rErr := GetError(RestErrorRPM, msg)
		//re.l.Warn(rErr.Error())
		return rErr
	}
	return nil
}

func (re *RestEndpoint) GetAddress() (string, int) {
	return re.rec.Addrs[0], re.rec.Port
}


///////////////////////////////////////////////////////////////////////////////
// Volumes


func (s *RestEndpoint) GetVolume(ctx context.Context, pool string, vname string) (*ResourceVolume, RestError) {

	var resvol ResourceVolume
	var rsp = GeneralResponse{Data: &resvol}

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", pool, vname)
	
	l := s.l.WithFields(log.Fields{
		"func": "GetVolume",
		"traceId": ctx.Value("traceId"),
		"url": addr,
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

	return  nil, getError(ctx, body)
}

func (s *RestEndpoint) CreateVolume(ctx context.Context, pool string, vol CreateVolumeDescriptor) RestError {
	
	addr := fmt.Sprintf("api/v3/pools/%s/volumes", pool)
	
	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "CreateVolume",
	        "traceId": ctx.Value("traceId"),
		"url": addr,

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
		"func": "DeleteVolume",
		"url": addr,
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

func (s *RestEndpoint) ListVolumes(ctx context.Context, pool string, vols *[]ResourceVolume) (RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/volumes", pool)

	l := s.l.WithFields(log.Fields{
		"func": "ListVolumes",
		"traceId": ctx.Value("traceId"),
	})

	l.Debug("Listing volumes")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetVolumesRCode)

	if err != nil {
		l.Warnln("Unable to list volumes", err.Error)
		return err
	}
	
	var rsp = &GetVolumesData{Data : vols }

	if stat == CodeOK {
		l.Debug("Obtained volume listing")
		return s.unmarshal(body, &rsp) 
	}

	return getError(ctx, body)

	// if err := ; err != nil {
	// 	l.Debugf("Unmarshaling error ", err.Error())
	// 	return err
	// }
	// l.Debugln("List data %+v", *rsp.Data)
	//return nil

}

// GetVolumeSnapshot provides information about specific volume snapshot requested
func (s *RestEndpoint) GetVolumeSnapshot(ctx context.Context, pool string, vname string, sname string) (sdp *ResourceSnapshot, err RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s", pool, vname, sname)

	l := jcom.LFC(ctx)

	l = s.l.WithFields(log.Fields{
		"section": "rest",
		"func": "GetVolumeSnapshot",
		"url": addr,
	})

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetSnapshotRCode)

	if err != nil {
		l.Warnf("Unable to get volume snapshots %+v", err.Error)
		return nil, err
	}

	var snapdata ResourceSnapshot
	var rsp = &GeneralResponse{Data : &snapdata }

	if stat == CodeOK || stat == CodeAccepted {
		l.Debug("Obtained volume listing")
		if err = s.unmarshal(body, &rsp); err != nil {
			return nil, err
		}
		return &snapdata, nil
	}

	return nil, getError(ctx, body)
}

// # Create Snapshot from existing volumes
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


func (s *RestEndpoint) getResultsEntries(data *GeneralResponse) (results *int64, entries interface{}) {
	re, ok := data.Data.(ResultEntries)
	if ok {
		return &re.Results, re.Entries
	}
	return nil, nil
}




//func (s *RestEndpoint) DeleteClone(
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

func (s *RestEndpoint) PromoteClone(vname string, sname string, cname string) RestError {

	return nil
	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "PromoteClone",
	// })

	// // Promote
	// addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones/%s/promote", s.pool, vname, sname, cname)
	// msg := fmt.Sprintf("Promoting clone %s", cname)
	// l.Trace(msg)
	// stat, body, err := s.rp.Send("POST", addr, nil, CreateTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// if stat == PromoteCloneRCode {
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

	// default:
	// 	msg := fmt.Sprintf("Unknown error %d, %s",
	// 		(*errData).Errno,
	// 		(*errData).Message)
	// 	s.l.Warn(msg)
	// 	return GetError(RestStorageFailureUnknown, msg)

	// }

	// return nil
}

func (s *RestEndpoint) GetTarget(tname string) (*Target, RestError) {
	return nil, nil

	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "GetTarget",
	// })
	// addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", s.pool, tname)
	// l.Tracef(fmt.Sprintf("Getting target %s information", tname))
	// stat, body, err := s.rp.Send("GET", addr, nil, GetTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return nil, GetError(RestRequestMalfunction, msg)
	// }

	// switch stat {
	// case GetTargetRCode:
	// case GetTargetRCodeDoNotExists:
	// 	return nil, GetError(RestResourceDNE, addr)
	// default:
	// 	return nil, GetError(RestStorageFailureUnknown, addr)

	// }

	// var rsp = &GetTargetData{}
	// if errC := json.Unmarshal(body, &rsp); errC != nil {
	// 	msg := fmt.Sprintf(
	// 		"%s Data: %s, Err: %+v.",
	// 		addr,
	// 		string(body[:len(body)]),
	// 		errC)
	// 	return nil, GetError(RestRPM, msg)
	// }

	// return &rsp.Data, nil
}

func (s *RestEndpoint) CreateTarget(tname string) RestError {

	return nil

	// tname = strings.ToLower(tname)

	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "CreateTarget",
	// })

	// data := CreateTarget{
	// 	Name:                tname,
	// 	Active:              true,
	// 	IncomingUsersActive: true,
	// }

	// addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets", s.pool)

	// l.Trace(fmt.Sprintf("Creating targets for volume: %s", tname))
	// stat, body, err := s.rp.Send("POST", addr, data, CreateTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// // Request is OK, exiting
	// if stat == CreateTargetRCode {
	// 	return nil
	// }

	// // Extract error information
	// if body == nil {
	// 	msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
	// 	l.Warn(msg)
	// 	return GetError(RestFailureUnknown, msg)
	// }

	// errData, er := s.getError(body)

	// if er != nil {
	// 	msg := fmt.Sprintf("Unable to extract err message %+v", er)
	// 	s.l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// switch (*errData).Errno {

	// default:
	// 	msg := fmt.Sprintf("Unknown error %d, %s",
	// 		(*errData).Errno,
	// 		(*errData).Message)
	// 	s.l.Warn(msg)
	// 	return GetError(RestStorageFailureUnknown, msg)

	// }

}

func (s *RestEndpoint) DeleteTarget(tname string) RestError {

	return nil

	//tname = strings.ToLower(tname)
	//l := s.l.WithFields(logrus.Fields{
	//	"func": "CreateTarget",
	//})

	//addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", s.pool, tname)

	//l.Trace(fmt.Sprintf("Deleating targets for volume: %s", tname))
	//stat, body, err := s.rp.Send("DELETE", addr, nil, DeleteTargetRCode)

	//if err != nil {
	//	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	//	l.Warn(msg)
	//	return GetError(RestRequestMalfunction, msg)
	//}

	//// Request is OK, exiting
	//if stat == DeleteTargetRCode {
	//	return nil
	//}

	//// Request is OK, exiting
	//if stat == 404 {
	//	msg := fmt.Sprintf("Target do not exists %s", tname)
	//	s.l.Warn(msg)
	//	return GetError(RestResourceDNE, msg)
	//}

	//// Extract error information
	//if body == nil {
	//	msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
	//	l.Warn(msg)
	//	return GetError(RestFailureUnknown, msg)
	//}

	//errData, er := s.getError(body)

	//if er != nil {
	//	msg := fmt.Sprintf("Unable to extract err message %+v", er)
	//	s.l.Warn(msg)
	//	return GetError(RestRequestMalfunction, msg)
	//}

	//switch (*errData).Errno {

	//default:
	//	msg := fmt.Sprintf("Unknown error %d, %s",
	//		(*errData).Errno,
	//		(*errData).Message)
	//	s.l.Warn(msg)
	//	return GetError(RestStorageFailureUnknown, msg)

	//}

}

func (s *RestEndpoint) AttachToTarget(tname string,
	vname string,
	mode string) RestError {
	return nil

	// tname = strings.ToLower(tname)

	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "AttachToTarget",
	// })

	// data := AttachToTarget{
	// 	Name: vname,
	// 	Lun:  0,
	// 	Mode: "wt",
	// }

	// addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns", s.pool, tname)

	// l.Tracef(fmt.Sprintf("Attaching volume to target: %s", tname))
	// stat, body, err := s.rp.Send("POST", addr, data, AttachToTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// // Request is OK, exiting
	// if stat == AttachToTargetRCode {
	// 	return nil
	// }

	// // Extract error information
	// if body == nil {
	// 	msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
	// 	l.Warn(msg)
	// 	return GetError(RestFailureUnknown, msg)
	// }

	// errData, er := s.getError(body)

	// if er != nil {
	// 	msg := fmt.Sprintf("Unable to extract err message %s", er)
	// 	s.l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// switch (*errData).Errno {

	// default:
	// 	msg := fmt.Sprintf("Unknown error %d, %s",
	// 		(*errData).Errno,
	// 		(*errData).Message)
	// 	s.l.Warn(msg)
	// 	return GetError(RestStorageFailureUnknown, msg)

	// }

}

func (s *RestEndpoint) DettachFromTarget(tname string, vname string) RestError {
	
	return nil

	// tname = strings.ToLower(tname)

	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "DettachFromTarget",
	// })

	// addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns/%s", s.pool, tname, vname)

	// msg := fmt.Sprintf("Detach volume from target: %s", tname)
	// l.Tracef(msg)
	// stat, body, err := s.rp.Send("DELETE", addr, nil, AttachToTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// // Request is OK, exiting
	// if stat == DettachFromTargetRCode {
	// 	return nil
	// }

	// // Request is OK, exiting
	// if stat == 404 {
	// 	msg := fmt.Sprintf("Target do not exists %s", vname)
	// 	s.l.Warn(msg)
	// 	return GetError(RestResourceDNE, msg)
	// }

	// // Extract error information
	// if body == nil {
	// 	msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
	// 	l.Warn(msg)
	// 	return GetError(RestFailureUnknown, msg)
	// }

	// errData, er := s.getError(body)

	// if er != nil {
	// 	msg := fmt.Sprintf("Unable to extract err message %+v", er)
	// 	s.l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// switch (*errData).Errno {

	// default:
	// 	msg := fmt.Sprintf("Unknown error %d, %s",
	// 		(*errData).Errno,
	// 		(*errData).Message)
	// 	s.l.Warn(msg)
	// 	return GetError(RestStorageFailureUnknown, msg)

	// }

}

func (s *RestEndpoint) AddUserToTarget(tname string,
	name string,
	pass string) RestError {

	return nil

	// tname = strings.ToLower(tname)

	// l := s.l.WithFields(logrus.Fields{
	// 	"func": "AddUserToTarget",
	// })

	// data := AddUserToTarget{
	// 	Name:     name,
	// 	Password: pass,
	// }

	// addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/incoming-users", s.pool, tname)

	// l.Tracef("Set CHAP user for tartget: %s", tname)
	// stat, body, err := s.rp.Send("POST", addr, data, AddUserToTargetRCode)

	// if err != nil {
	// 	msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
	// 	l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// // Request is OK, exiting
	// if stat == AddUserToTargetRCode {
	// 	return nil
	// }

	// // Extract error information
	// if body == nil {
	// 	msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
	// 	l.Warn(msg)
	// 	return GetError(RestFailureUnknown, msg)
	// }

	// errData, er := s.getError(body)

	// if er != nil {
	// 	msg := fmt.Sprintf("Unable to extract err message %s", er)
	// 	s.l.Warn(msg)
	// 	return GetError(RestRequestMalfunction, msg)
	// }

	// switch (*errData).Errno {

	// default:
	// 	msg := fmt.Sprintf("Unknown error %d, %s",
	// 		(*errData).Errno,
	// 		(*errData).Message)
	// 	s.l.Warn(msg)
	// 	return GetError(RestStorageFailureUnknown, msg)

	// }

}

func GetTimeStamp(tRaw string) (int64, RestError) {
	layout := "2006-1-2 15:4:5"
	t, err := time.Parse(layout, tRaw)
	if err != nil {
		msg := fmt.Sprintf("Unable to extract time stamp: %s", err)
		return 0, GetError(RestErrorRequestMalfunction, msg)
	}
	return t.Unix(), nil
}
