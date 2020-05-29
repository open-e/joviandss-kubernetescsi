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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateVolumeDescriptor struct {
	Name string
	Size int64
}

type SnapshotDescriptor struct {
	VName   string
	SName   string
	Created string
}

func (s *Storage) getError(body []byte) (out *ErrorT, err error) {
	var edata ErrorData
	if err := json.Unmarshal(body, &edata); err != nil {
		bs := fmt.Sprintf(string(body[:len(body)]))
		msg := fmt.Sprintf("Unable to extract json output from error message: %s", bs)
		err = status.Error(codes.Internal, msg)
		s.l.Warnf(msg)
		return nil, err
	}

	if edata.Error.Errno == 0 {
		bs := fmt.Sprintf(string(body[:len(body)]))
		msg := fmt.Sprintf("Error number was not set: %s", bs)
		err = status.Error(codes.Internal, msg)
		s.l.Warnf(msg)
		return nil, err

	}

	if len(edata.Error.Message) == 0 {
		bs := fmt.Sprintf(string(body[:len(body)]))
		msg := fmt.Sprintf("Error message was not set: %s", bs)
		err = status.Error(codes.Internal, msg)
		s.l.Warnf(msg)
		return nil, err
	}

	return &edata.Error, nil
}

func (s *Storage) GetAddress() (string, int) {
	return s.addr, s.port

}

///////////////////////////////////////////////////////////////////////////////
// Pools

func (s *Storage) GetPool() (*Pool, RestError) {
	l := s.l.WithFields(logrus.Fields{
		"func": "GetPool",
	})
	msg := fmt.Sprintf("Getting pool %s", s.pool)
	l.Trace(msg)
	addr := fmt.Sprintf("api/v3/pools/%s", s.pool)
	stat, body, err := s.rp.Send("GET", addr, nil, GetPoolRCode)
	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	switch stat {
	case 500:
		return nil, GetError(RestResourceDNE, addr)
	case GetPoolRCode:
	default:
		return nil, GetError(RestFailureUnknown, addr)
	}

	var poolData GetPoolData
	if err := json.Unmarshal(body, &poolData); err != nil {
		return nil, GetError(RestRPM, fmt.Sprintf("Error %s for %s", err.Error(), string(body[:len(body)])))
	}

	return &poolData.Data, nil
}

// GetPools return list of pools
func (s *Storage) GetPools() ([]PoolShort, error) {
	l := s.l.WithFields(logrus.Fields{
		"func": "GetPools",
	})
	var ps = []PoolShort{}
	stat, body, err := s.rp.Send("GET", "api/v3/pools", nil, GetPoolsRCode)
	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	if stat != GetPoolsRCode {
		return nil, err
	}

	var rsp GetPoolsData
	if err := json.Unmarshal(body, &rsp); err != nil {
		panic(err)
	}
	for poolN := range rsp.Data {
		ps = append(ps, rsp.Data[poolN])
	} //fmt.Println("Body %+v", body)
	return ps, nil
}

///////////////////////////////////////////////////////////////////////////////
// Volumes

func (s *Storage) VolumeExists(vname string) (bool, error) {
	l := s.l.WithFields(logrus.Fields{
		"func": "VolumeExists",
	})

	l.Trace("Get Existing volumes")
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", s.pool, vname)

	stat, _, _ := s.rp.Send("GET", addr, nil, GetVolumeRCode)

	if stat == GetVolumeRCode {
		return true, nil
	}
	return false, nil
}

func (s *Storage) GetVolume(vname string) (*Volume, RestError) {

	l := s.l.WithFields(logrus.Fields{
		"func": "GetVolume",
	})

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", s.pool, vname)

	stat, body, err := s.rp.Send("GET", addr, nil, GetVolumeRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	if stat != GetVolumeRCode {
		errData, errC := s.getError(body)
		if errC != nil {
			msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
			rErr := GetError(RestRPM, msg)
			l.Warn(rErr.Error())
			return nil, rErr

		}
		msg := fmt.Sprintf("Internal failure during obtaining volume info, error: %s.", (*errData).Message)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)

	}

	var rsp = &GetVolumeData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
		rErr := GetError(RestRPM, msg)

		l.Warn(rErr.Error())
		return nil, rErr

	}

	return &rsp.Data, nil

}

func (s *Storage) CreateVolume(vdesc CreateVolumeDescriptor) RestError {
	data := CreateVolume{
		Name: vdesc.Name,
		Size: fmt.Sprintf("%d", vdesc.Size)}

	addr := fmt.Sprintf("api/v3/pools/%s/volumes", s.pool)

	stat, body, err := s.rp.Send("POST", addr, data, CreateVolumeRCode)

	if stat == CreateVolumeRCode {
		return nil
	}

	if err != nil {
		msg := fmt.Sprintf("Unable to process rest request: %s", err.Error())
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)

	}

	s.l.Tracef(fmt.Sprintf("Unable to create volume: %s", string(body[:len(body)])))

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {
	case CreateVolumeECodeExists:
		msg := fmt.Sprintf("Volume %s already exists exist", vdesc.Name)
		s.l.Warn(msg)
		return GetError(RestObjectExists, msg)

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

	return nil
}

// DeleteVolume delete volume, fails if it has snapshots
//
// set rSnapshots to true in order to delete snapshots
func (s *Storage) DeleteVolume(vname string, rSnapshots bool) RestError {
	var err error
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s", s.pool, vname)

	data := DeleteVolume{
		RecursivelyChildren:   false,
		RecursivelyDependents: false,
		ForceUmount:           false,
	}
	if rSnapshots == true {
		data.RecursivelyChildren = true
	}

	stat, body, err := s.rp.Send("DELETE", addr, data, DeleteVolumeRCode)

	s.l.Tracef("Status %d", stat)
	s.l.Tracef("Body %s", body[:len(body)])
	s.l.Tracef("Err %+v", err)

	if stat == DeleteVolumeRCode {
		return nil
	}

	if err != nil {
		msg := fmt.Sprintf("Internal failure during volume %s deletion.", vname)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	if (*errData).Errno == 1 {
		msg := fmt.Sprintf("Volume %s doesn't exist", vname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, "")
	}

	if (*errData).Errno == 1000 {

		msg := fmt.Sprintf("Volume %s is busy", vname)
		s.l.Warn(msg)
		return GetError(RestResourceBusy, msg)
	}

	msg := fmt.Sprintf("Unidentified failure during volume %s deletion.", vname)
	s.l.Warn(msg)
	return GetError(RestFailureUnknown, msg)
}

func (s *Storage) ListVolumes() ([]string, RestError) {
	var err error
	addr := fmt.Sprintf("api/v3/pools/%s/volumes", s.pool)

	stat, body, err := s.rp.Send("GET", addr, nil, GetVolumesRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		s.l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	var rsp = &GetVolumesData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
		rErr := GetError(RestRPM, msg)

		s.l.Warn(rErr.Error())
		return nil, rErr

	}

	if stat != GetVolumesRCode {
		errData, errC := s.getError(body)
		if errC != nil {
			msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
			rErr := GetError(RestRPM, msg)

			s.l.Warn(rErr.Error())
			return nil, rErr

		}
		msg := fmt.Sprintf("Internal failure during volume listing, error: %+v.", errData.Message)
		s.l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	vl := make([]string, len(rsp.Data))
	for i, v := range rsp.Data {
		vl[i] = v.Name
	}

	return vl, nil
}

func (s *Storage) GetSnapshot(vname string, sname string) (*Snapshot, RestError) {

	l := s.l.WithFields(logrus.Fields{
		"func": "GetVolume",
	})

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s",
		s.pool, vname, sname)

	stat, body, err := s.rp.Send("GET", addr, nil, GetSnapshotRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	switch stat {

	case GetSnapshotRCodeDNE:

		msg := fmt.Sprintf("Snapshot %s does not exist.", sname)
		errData, errC := s.getError(body)
		if errC == nil {
			msg += fmt.Sprintf("Err: %s.", (*errData).Message)
		}

		l.Warn(msg)
		return nil, GetError(RestResourceDNE, msg)

	case GetSnapshotRCode:

	default:
		errData, errC := s.getError(body)
		if errC != nil {
			msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
			rErr := GetError(RestRPM, msg)
			l.Warn(rErr.Error())
			return nil, rErr

		}
		msg := fmt.Sprintf("Internal failure during obtaining volume info, error: %s", errData.Message)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	var rsp = &GetSnapshotData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf("Data: %s, Err: %+v.", string(body[:len(body)]), errC)
		rErr := GetError(RestRPM, msg)

		l.Warn(rErr.Error())
		return nil, rErr
	}

	return &rsp.Data, nil

}

func (s *Storage) CreateSnapshot(vname string, sname string) RestError {

	l := s.l.WithFields(logrus.Fields{
		"func": "CreateSnapshot",
	})

	data := CreateSnapshot{
		Snapshot_name: sname}

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots", s.pool, vname)

	stat, body, err := s.rp.Send("POST", addr, data, CreateSnapshotRCode)

	if err != nil {
		return GetError(RestRequestMalfunction, addr)
	}

	// Request is OK, exiting
	if stat == CreateSnapshotRCode {
		return nil
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {
	case 1:
		msg := fmt.Sprintf("Snapshot %s doesn't exist", vname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)
	case CreateSnapshotECodeExists:
		msg := fmt.Sprintf("Snapshot %s already exists", sname)
		s.l.Warn(msg)
		return GetError(RestObjectExists, msg)

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) DeleteSnapshot(vname string, sname string) RestError {
	var err error
	l := s.l.WithFields(logrus.Fields{
		"func": "DeleteSnapshot",
	})
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s", s.pool, vname, sname)

	data := DeleteSnapshot{
		Recursively_dependents: false,
	}

	stat, body, err := s.rp.Send("DELETE", addr, data, DeleteSnapshotRCode)

	if err != nil {
		return GetError(RestRequestMalfunction, addr)
	}

	// Request is OK, exiting
	if stat == DeleteSnapshotRCode {
		return nil
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {
	case 1:
		msg := fmt.Sprintf("Snapshot %s doesn't exist", sname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)

	case 1000:
		msg := fmt.Sprintf("Snapshot %s is busy. Msg: %s ", sname, (*errData).Message)
		s.l.Warn(msg)

		return GetError(RestResourceBusy, msg)

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) ListAllSnapshots(f func(string) bool) ([]SnapshotShort, RestError) {
	var err error
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/snapshots", s.pool)

	stat, body, err := s.rp.Send("GET", addr, nil, GetAllSnapshotsRCode)

	if err != nil {
		return nil, GetError(RestRequestMalfunction, addr)
	}

	// error handling
	if stat != GetAllSnapshotsRCode {
		var errData *ErrorT
		var errC error
		if errData, errC = s.getError(body); errC != nil {
			return nil, GetError(RestRPM, string(body[:len(body)]))
		}

		msg := fmt.Sprintf(
			"request to %s failed with %+v",
			addr,
			*errData)
		return nil, GetError(RestStorageFailureUnknown, msg)
	}

	// data extraction
	var rsp = &GetAllSnapshotsData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf(
			"Unable to unmarshal: %s, Err: %+v.",
			string(body[:len(body)]),
			errC)
		rErr := GetError(RestRPM, msg)
		return nil, rErr
	}
	// no snapshots
	if rsp.Data.Results <= 0 {
		return nil, nil
	}

	var out []SnapshotShort
	s.l.Debugf("Entries: %+v", rsp.Data.Entries)

	for _, se := range rsp.Data.Entries {
		s.l.Debugf("Snap: %+v", se)
		if f != nil {
			if !f(se.Name) {
				continue
			}
		}
		out = append(out, se)
	}

	return out, nil
}

func (s *Storage) ListVolumeSnapshots(vname string, f func(string) bool) ([]SnapshotShort, RestError) {
	var err error
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots", s.pool, vname)

	stat, body, err := s.rp.Send("GET", addr, nil, GetVolSnapshotsRCode)
	if err != nil {
		return nil, GetError(RestRequestMalfunction, addr)
	}

	if stat != GetVolSnapshotsRCode {
		errData, errC := s.getError(body)
		if errC != nil {
			msg := GetError(
				RestRPM,
				fmt.Sprintf(
					"%s, Data %s, Err %s",
					addr,
					string(body[:len(body)]),
					errC))

			return nil, msg
		}
		if errData.Class == "zfslib.zfsapi.resources.ZfsResourceError" {
			return nil, GetError(RestResourceDNE, addr)
		}
		msg := fmt.Sprintf("%s, Data %s", addr, errData.Message)
		return nil, GetError(RestStorageFailureUnknown, msg)
	}

	var rsp = &GetVolSnapshotsData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf(
			"%s Data: %s, Err: %+v.",
			addr,
			string(body[:len(body)]),
			errC)
		return nil, GetError(RestRPM, msg)
	}

	// prepare response data
	var i int64
	i = 0

	if rsp.Data.Results <= 0 {
		return nil, nil
	}

	var out []SnapshotShort
	for _, se := range rsp.Data.Entries {

		if f != nil {
			if !f(se.Name) {
				continue
			}
		}
		timeStamp, r_err := GetTimeStamp(se.Creation)
		if r_err != nil {
			continue
		}
		ss := SnapshotShort{
			Volume:     vname,
			Name:       se.Name,
			Clones:     se.Clones,
			Properties: SnapshotProperties{strconv.FormatInt(timeStamp, 10)},
		}
		out = append(out, ss)
		i += 1
	}

	return out, nil
}

// CreateClone creates clone of a volume from snapshot
func (s *Storage) CreateClone(vname string, sname string, cname string) RestError {
	l := s.l.WithFields(logrus.Fields{
		"func": "CreateClone",
	})

	// Clone

	data := CreateClone{
		Name:     cname,
		Snapshot: sname,
	}
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/clone", s.pool, vname)
	msg := fmt.Sprintf("Creating clone of snapshot %s  volume: %s", sname, vname)
	l.Trace(msg)
	stat, body, err := s.rp.Send("POST", addr, data, CreateCloneRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	if stat == CreateCloneRCode {
		return nil
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	case 1:
		msg := fmt.Sprintf("Clone %s doesn't exist", cname)
		return GetError(RestResourceDNE, msg)
	case 100:
		msg := fmt.Sprintf("Target volume %s already exists", cname)
		return GetError(RestObjectExists, msg)
	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

	return nil
}

func (s *Storage) DeleteClone(
	vname string,
	sname string,
	cname string,
	rChildren bool,
	rDependent bool) RestError {

	l := s.l.WithFields(logrus.Fields{
		"func": "Delete Clone of the Volume",
	})

	data := DeleteClone{
		RecursivelyChildren:   rChildren,
		RecursivelyDependents: rDependent,
		ForceUmount:           false,
	}
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones/%s", s.pool, vname, sname, cname)
	msg := fmt.Sprintf("Deleting clone of snapshot %s  volume: %s", sname, vname)
	l.Trace(msg)
	stat, body, err := s.rp.Send("DELETE", addr, data, CreateCloneRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	if stat == DeleteCloneRCode {
		return nil
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	case 1:
		msg := fmt.Sprintf("Clone %s doesn't exist", cname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)
	case 1000:
		msg := fmt.Sprintf("Clone %s may have snapshots", cname)
		s.l.Warn(msg)
		return GetError(RestResourceBusy, msg)
	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

	return nil
}

func (s *Storage) PromoteClone(vname string, sname string, cname string) RestError {

	l := s.l.WithFields(logrus.Fields{
		"func": "PromoteClone",
	})

	// Promote
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones/%s/promote", s.pool, vname, sname, cname)
	msg := fmt.Sprintf("Promoting clone %s", cname)
	l.Trace(msg)
	stat, body, err := s.rp.Send("POST", addr, nil, CreateTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	if stat == PromoteCloneRCode {
		return nil
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {
	case 1:
		msg := fmt.Sprintf("Clone %s doesn't exist", cname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

	return nil
}

func (s *Storage) GetTarget(tname string) (*Target, RestError) {
	l := s.l.WithFields(logrus.Fields{
		"func": "GetTarget",
	})
	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", s.pool, tname)
	l.Tracef(fmt.Sprintf("Getting target %s information", tname))
	stat, body, err := s.rp.Send("GET", addr, nil, GetTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	switch stat {
	case GetTargetRCode:
	case GetTargetRCodeDoNotExists:
		return nil, GetError(RestResourceDNE, addr)
	default:
		return nil, GetError(RestStorageFailureUnknown, addr)

	}

	var rsp = &GetTargetData{}
	if errC := json.Unmarshal(body, &rsp); errC != nil {
		msg := fmt.Sprintf(
			"%s Data: %s, Err: %+v.",
			addr,
			string(body[:len(body)]),
			errC)
		return nil, GetError(RestRPM, msg)
	}

	return &rsp.Data, nil
}

func (s *Storage) CreateTarget(tname string) RestError {

	tname = strings.ToLower(tname)

	l := s.l.WithFields(logrus.Fields{
		"func": "CreateTarget",
	})

	data := CreateTarget{
		Name:                tname,
		Active:              true,
		IncomingUsersActive: true,
	}

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets", s.pool)

	l.Trace(fmt.Sprintf("Creating targets for volume: %s", tname))
	stat, body, err := s.rp.Send("POST", addr, data, CreateTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	// Request is OK, exiting
	if stat == CreateTargetRCode {
		return nil
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) DeleteTarget(tname string) RestError {
	tname = strings.ToLower(tname)
	l := s.l.WithFields(logrus.Fields{
		"func": "CreateTarget",
	})

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", s.pool, tname)

	l.Trace(fmt.Sprintf("Deleating targets for volume: %s", tname))
	stat, body, err := s.rp.Send("DELETE", addr, nil, DeleteTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	// Request is OK, exiting
	if stat == DeleteTargetRCode {
		return nil
	}

	// Request is OK, exiting
	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) AttachToTarget(tname string,
	vname string,
	mode string) RestError {
	tname = strings.ToLower(tname)

	l := s.l.WithFields(logrus.Fields{
		"func": "AttachToTarget",
	})

	data := AttachToTarget{
		Name: vname,
		Lun:  0,
		Mode: "wt",
	}

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns", s.pool, tname)

	l.Tracef(fmt.Sprintf("Attaching volume to target: %s", tname))
	stat, body, err := s.rp.Send("POST", addr, data, AttachToTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	// Request is OK, exiting
	if stat == AttachToTargetRCode {
		return nil
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %s", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) DettachFromTarget(tname string, vname string) RestError {

	tname = strings.ToLower(tname)

	l := s.l.WithFields(logrus.Fields{
		"func": "DettachFromTarget",
	})

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns/%s", s.pool, tname, vname)

	msg := fmt.Sprintf("Detach volume from target: %s", tname)
	l.Tracef(msg)
	stat, body, err := s.rp.Send("DELETE", addr, nil, AttachToTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	// Request is OK, exiting
	if stat == DettachFromTargetRCode {
		return nil
	}

	// Request is OK, exiting
	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", vname)
		s.l.Warn(msg)
		return GetError(RestResourceDNE, msg)
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %+v", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func (s *Storage) AddUserToTarget(tname string,
	name string,
	pass string) RestError {
	tname = strings.ToLower(tname)

	l := s.l.WithFields(logrus.Fields{
		"func": "AddUserToTarget",
	})

	data := AddUserToTarget{
		Name:     name,
		Password: pass,
	}

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/incoming-users", s.pool, tname)

	l.Tracef("Set CHAP user for tartget: %s", tname)
	stat, body, err := s.rp.Send("POST", addr, data, AddUserToTargetRCode)

	if err != nil {
		msg := fmt.Sprintf("Internal failure in communication with storage %s.", s.addr)
		l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	// Request is OK, exiting
	if stat == AddUserToTargetRCode {
		return nil
	}

	// Extract error information
	if body == nil {
		msg := fmt.Sprintf("Unidentifiable error, code : %d.", stat)
		l.Warn(msg)
		return GetError(RestFailureUnknown, msg)
	}

	errData, er := s.getError(body)

	if er != nil {
		msg := fmt.Sprintf("Unable to extract err message %s", er)
		s.l.Warn(msg)
		return GetError(RestRequestMalfunction, msg)
	}

	switch (*errData).Errno {

	default:
		msg := fmt.Sprintf("Unknown error %d, %s",
			(*errData).Errno,
			(*errData).Message)
		s.l.Warn(msg)
		return GetError(RestStorageFailureUnknown, msg)

	}

}

func GetTimeStamp(tRaw string) (int64, RestError) {
	layout := "2006-1-2 15:4:5"
	t, err := time.Parse(layout, tRaw)
	if err != nil {
		msg := fmt.Sprintf("Unable to extract time stamp: %s", err)
		return 0, GetError(RestRequestMalfunction, msg)
	}
	return t.Unix(), nil
}
