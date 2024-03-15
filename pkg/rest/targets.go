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
	"fmt"
	//"reflect"
	//"strconv"
	//"strings"

	log "github.com/sirupsen/logrus"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	jcom "joviandss-kubernetescsi/pkg/common"
)

func (s *RestEndpoint) GetTarget(ctx context.Context, pool string, tname string) (*ResourceTarget, RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
		"section": "rest",
		"func": "GetTarget",
	})

	var outUser CreateTargetOutgoingUser
	var resTarget ResourceTarget
	resTarget.OutgoingUser = &outUser
	var rsp = GeneralResponse{Data: &resTarget}

	l.Debugf(fmt.Sprintf("Getting target %s information", tname))
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, CodeOK)

	if err != nil {
		msg := fmt.Sprintf("Unable to get target %s information", tname)
		l.Warn(msg)
		return nil, GetError(RestErrorRequestMalfunction, msg)
	}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}
	
	if stat == CodeOK {
		return &resTarget, nil
	}

	return nil, getError(ctx, body)
}



func (s *RestEndpoint) CreateTarget(ctx context.Context, pool string, desc *CreateTargetDescriptor) RestError {

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets", pool)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
		"section": "rest",
		"func": "CreateTarget",
	})

	l.Debugf(fmt.Sprintf("Creating target %s", desc.Name))
	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CodeCreated)

	if err != nil {
		l.Warnf("Unable to create target %s because of %s", desc.Name, err.Error)
		return err
	}

	if stat == CodeOK || stat == CodeAccepted {
		l.Debugf("Target %s created", desc.Name)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DeleteTarget(ctx context.Context, pool string, tname string) RestError {
	
	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
		"section": "rest",
		"func": "DeleteTarget",
	})

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, DeleteTargetRCode)
	
	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debug(msg)
		return GetError(RestErrorResourceDNE, msg)
	}
	
	if err != nil {
		l.Warnf("Unable to delete target %s because of %s", tname, err.Error)
		return err
	}

	if stat == CodeNoContent{
		l.Debugf("Target %s deleted", tname)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) AttachVolumeToTarget(ctx context.Context, pool string, tname string, desc *TargetLunDescriptor) (err RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
		"section": "rest",
		"func": "AttachVolumeToTarget",
	})

	l.Debugf("Attach volume with id %s to target: %s", desc.Name, tname)
	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CodeOK)
	
	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debugf(msg)
		return GetError(RestErrorResourceDNETarget, msg)
	}

	if err != nil {
		l.Warnf("Unable to attach volume with id %s to target %s because of %s", desc.Name, tname, err.Error)
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("Target %s created", desc.Name)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DettachVolumeFromTarget(ctx context.Context, pool string, tname string, vname string) RestError {
	
	addr := fmt.Sprintf("api/v3/pools/%s/san/iscsi/targets/%s/luns/%s", pool, tname, vname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url": addr,
		"section": "rest",
		"func": "DetachVolumeFromTarget",
	})

	l.Debugf("Detach volume %s from target: %s", vname, tname)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, CodeNoContent)
	
	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debugf(msg)
		return GetError(RestErrorResourceDNETarget, msg)
	}
	
	if err != nil {
		l.Warnln("Unable to detach volume with id %s from target %s because of %s", vname, tname, err.Error)
		return err
	}

	// Request is OK, exiting
	if stat == CodeNoContent {
		return nil
	}
	// TODO: provide erroro handling for various cases	
	return getError(ctx, body)
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

