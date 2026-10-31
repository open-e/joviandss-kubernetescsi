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
	//"reflect"
	//"strconv"
	//"strings"

	log "github.com/sirupsen/logrus"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// # Create Snapshot from existing NAS Volume
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing nas volume described with *vid*
//   - *vid* physical nas volume id as it used by JovianDSS
//   - *desc* data, including snapahot name, that would be transfered to create snapshot
func (s *RestEndpoint) CreateNASSnapshot(ctx context.Context, pool string, nvid string, desc *CreateNASSnapshotDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/pools/%s/nas-volumes/%s/snapshots", pool, nvid)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "CreateNASSnapshot",
		"section": "rest",
		"url":     addr,
	})
	ctx = jcom.WithLogger(ctx, l)

	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, CodeCreated)
	if err != nil {
		s.l.Warnln("Unable to create snapshot ", desc.Name)
		return err
	}

	if stat == CodeOK || stat == CodeCreated {
		l.Debugf("CreateSnapshot %s Done", desc.Name)
		return nil
	}

	return getError(ctx, body)
}

func (s *RestEndpoint) DeleteNASSnapshot(ctx context.Context, pool string, nvid string, sname string) (err RestError) {
	l := jcom.LFC(ctx)

	addr := fmt.Sprintf("api/v3/pools/%s/nas-volumes/%s/snapshots/%s", pool, nvid, sname)

	l = l.WithFields(log.Fields{
		"func":    "DeleteNASSnapshot",
		"section": "rest",
		"addr":    addr,
	})
	ctx = jcom.WithLogger(ctx, l)

	stat, body, err := s.rp.Send(ctx, "DELETE", addr, nil, CodeNoContent)
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
