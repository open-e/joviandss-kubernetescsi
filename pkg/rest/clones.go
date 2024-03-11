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

	jcom "joviandss-kubernetescsi/pkg/common"
)

func (s *RestEndpoint) GetVolumeSnapshotClones(ctx context.Context, pool string, vds string, sds string) (clones []ResourceVolumeSnapshotClones, err RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones", pool, vds, sds)
	
	l := s.l.WithFields(log.Fields{
		"func": "GetVolumeSnapshotClones",
		"section": "rest",
		"url": addr,
	})
	
	var rsp = GeneralResponse{Data: &clones}

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

	return  nil, getError(ctx, body)

}

// # CreateClone creates clone of a volume from snapshot
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *vid*
//   - *vid* physical volume id as it used by JovianDSS
//   - *desc* data, including new volume name, that would be transfered to create clone
func (s *RestEndpoint) CreateClone(ctx context.Context, pool string, vid string, desc CloneVolumeDescriptor) RestError {
	
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/clone", pool, vid)
	
	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "CreateClone",
		"url": addr,
		"section": "rest",
	})
	l.Debugf("Create clone $s from volume %s snapshot %s", vid, desc.Snapshot, desc.Name)

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


func (s *RestEndpoint) DeleteClone(ctx context.Context, pool string, vds string, sds string, cds string, desc DeleteVolumeDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/volumes/%s/snapshots/%s/clones/%s", pool, vds, sds, cds)
	
	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func": "DeleteClone",
		"url": addr,
		"section": "rest",
	})
	l.Debugf("Delete clone $s from volume %s snapshot %s", cds, vds, sds)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, desc, CreateCloneRCode)
	
	if err != nil {
		s.l.Warnln("Unable to delete clone ", cds)
		return err
	}

	if stat == CodeNoContent {
		return nil
	}
	
	return getError(ctx, body)
}
