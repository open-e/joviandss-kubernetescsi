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

package rest

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

func (s *RestEndpoint) GetNASVolumeSnapshotClones(ctx context.Context, pool string, vds string, sds string) (clones []ResourceVolumeSnapshotClones, err RestError) {
	addr := fmt.Sprintf("api/v3/pools/%s/nas-volumes/%s/snapshots/%s/clones", pool, vds, sds)

	l := s.l.WithFields(log.Fields{
		"func":    "GetNASVolumeSnapshotClones",
		"section": "rest",
		"url":     addr,
	})

	rsp := GeneralResponse{Data: &clones}

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetVolumeRCode)
	if err != nil {
		msg := fmt.Sprintf("Unable to get list of clones for snap %s of vol %s ", sds, vds)
		l.Warn(msg)
		return nil, GetError(RestErrorRequestMalfunction, msg)
	}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	if stat == CodeOK || stat == CodeNoContent {
		return clones, nil
	}

	return nil, getError(ctx, body)
}

// # CreateClone creates clone of a volume from snapshot
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *vid*
//   - *vid* physical volume id as it used by JovianDSS
//   - *desc* data, including new volume name, that would be transfered to create clone
func (s *RestEndpoint) CreateNASClone(ctx context.Context, pool string, nvid string, nsds string, desc CloneNASVolumeDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/nas-volumes/%s/snapshots/%s/clones", pool, nvid, nsds)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "CreateNASClone",
		"url":     addr,
		"section": "rest",
	})
	l.Debugf("Create clone %s from volume %s snapshot %s", desc.Name, nvid, nsds)

	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CreateCloneRCode)
	if err != nil {
		s.l.Warnln("Unable to create clone ", desc.Name)
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DeleteNASClone(ctx context.Context, pool string, nvds string, nsds string, ncds string) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/nas-volumes/%s/snapshots/%s/clones/%s", pool, nvds, nsds, ncds)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "DeleteNASClone",
		"url":     addr,
		"section": "rest",
	})
	l.Debugf("Delete clone %s from volume %s snapshot %s", ncds, nvds, nsds)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, CodeNoContent)
	if err != nil {
		s.l.Warnf("Unable to delete NAS clone %s", ncds)
		return err
	}

	if stat == CodeNoContent {
		return nil
	}

	return getError(ctx, body)
}
