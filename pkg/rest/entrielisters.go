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
	// "strings"
	// "regexp"
	// "runtime/debug"

	log "github.com/sirupsen/logrus"

	jcom "joviandss-kubernetescsi/pkg/common"
)

type GetResourceEntries func(ctx context.Context, pool string, page int64, dc int64, args interface{}) (ent *ResultEntries, err RestError)

func pagedcSuffix(addr string, page *int64, dc *int64) (out string) {
	
	if page != nil || dc != nil {
		addr += "?"
	}

	if page != nil {
		addr += fmt.Sprintf("page=%d", *page)
	}

	if dc != nil {
		if page != nil {
			addr += "&"
		}
		addr += fmt.Sprintf("_dc=%d", *dc)
	}

	return addr
}

func (s *RestEndpoint) GetVolumeSnapshotsEntries(ctx context.Context, pool string, vname string, page int64, dc int64) (ent ResultEntries, err RestError) {

	return ent, nil
}

func (s *RestEndpoint) GetVolumesEntries(ctx context.Context, pool string, page int64, dc int64) (ent ResultEntries, err RestError) {

	return ent, nil
}

func (s *RestEndpoint) GetSnapshotsEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {

	addr := fmt.Sprintf("api/v3/pools/%s/volumes/snapshots", pool)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func": "GetSnapshotsEntries",
		"addr": addr,
		"section": "rest",
	})

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetAllSnapshotsRCode)

	if err != nil {
		s.l.Warnf("Unable to get snapshot list for pool %s", pool)
		return  nil, err
	}

	var snaps []ResourceSnapshotShort
	var entries = ResultEntries{Entries: &snaps}
	var rsp = GeneralResponse{Data: &entries}
	//var rsp GeneralResponse
	//{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		if rsp.Data != nil {
			data, ok := rsp.Data.(*ResultEntries)

			if ok {
				return data, nil
			}
			return nil, GetError(RestErrorRequestMalfunction, fmt.Sprintf("response is not expected %+v", *data))
		}
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}
