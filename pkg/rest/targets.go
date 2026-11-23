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

	log "github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

func (s *RestEndpoint) GetTarget(ctx context.Context, pool string, tname string) (*ResourceTarget, RestError) {

	addr := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url":     addr,
		"section": "rest",
		"func":    "GetTarget",
	})

	var outUser CreateTargetOutgoingUser
	var resTarget ResourceTarget
	resTarget.OutgoingUser = &outUser
	var rsp = GeneralResponse{Data: &resTarget}

	l.Debugf("Getting target %s information", tname)
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

	addr := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets", pool)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url":     addr,
		"section": "rest",
		"func":    "CreateTarget",
	})

	l.Debugf("Creating target %s", desc.Name)
	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CodeCreated)

	if err != nil {
		l.Warnf("Unable to create target %s because of %s", desc.Name, err.Error())
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("Target %s created", desc.Name)
		return nil
	}

	if stat == CodeAccepted {
		var rErr RestError
		for i := 0; i < 10; i++ {
			_, rErr = s.GetTarget(ctx, pool, desc.Name)
			if rErr == nil {
				l.Debugf("Target %s created", desc.Name)
			}
		}
		return rErr
	}
	return getError(ctx, body)
}

func (s *RestEndpoint) DeleteTarget(ctx context.Context, pool string, tname string) RestError {

	addr := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url":     addr,
		"section": "rest",
		"func":    "DeleteTarget",
	})

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, DeleteTargetRCode)

	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debug(msg)
		return GetError(RestErrorResourceDNE, msg)
	}

	if err != nil {
		l.Warnf("Unable to delete target %s because of %s", tname, err.Error())
		return err
	}

	if stat == CodeNoContent {
		l.Debugf("Target %s deleted", tname)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) AttachVolumeToTarget(ctx context.Context, pool string, tname string, desc *TargetLunDescriptor) (err RestError) {

	addr := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s/luns", pool, tname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url":     addr,
		"section": "rest",
		"func":    "AttachVolumeToTarget",
	})

	l.Debugf("Attach volume with id %s to target: %s", desc.Name, tname)
	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CodeOK)

	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debugf("%s", msg)
		return getError(ctx, body)
	}

	if err != nil {
		l.Warnf("Unable to attach volume with id %s to target %s because of %s", desc.Name, tname, err.Error())
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("Target %s created", desc.Name)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DettachVolumeFromTarget(ctx context.Context, pool string, tname string, vname string) RestError {

	addr := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s/luns/%s", pool, tname, vname)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"url":     addr,
		"section": "rest",
		"func":    "DetachVolumeFromTarget",
	})

	l.Debugf("Detach volume %s from target: %s", vname, tname)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, CodeNoContent)

	if stat == 404 {
		msg := fmt.Sprintf("Target do not exists %s", tname)
		l.Debugf("%s", msg)
		return GetError(RestErrorResourceDNETarget, msg)
	}

	if err != nil {
		l.Warnf("Unable to detach volume with id %s from target %s because of %s", vname, tname, err.Error())
		return err
	}

	// Request is OK, exiting
	if stat == CodeNoContent {
		return nil
	}
	// TODO: provide erroro handling for various cases
	return getError(ctx, body)
}
