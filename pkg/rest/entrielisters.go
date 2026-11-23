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

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// ListPerPage is the page size sent with every list request. Without
// per_page an appliance returns pages of its own default size, and
// JovianDSS 2 ignores page and returns the first page again and again.
const ListPerPage = 100

func pagedcSuffix(addr string, page *int64, dc *int64) (out string) {
	if page != nil || dc != nil {
		addr += "?"
	}

	if page != nil {
		addr += fmt.Sprintf("page=%d&per_page=%d", *page, ListPerPage)
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

func (s *RestEndpoint) GetVolumesEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/volumes", pool)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "GetVolumeEntries",
		"addr":    addr,
		"section": "rest",
	})

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, CodeOK)
	if err != nil {
		s.l.Warnf("Unable to get snapshot list for pool %s", pool)
		return nil, err
	}

	var vols []ResourceVolume
	entries := ResultEntries{Entries: &vols}
	rsp := GeneralResponse{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		// each entry is decoded and checked for its required fields
		checked, results, errD := decodeEntries[ResourceVolume](body)
		if errD != nil {
			return nil, errD
		}
		return &ResultEntries{Results: results, Entries: &checked}, nil
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}

func (s *RestEndpoint) GetSnapshotsEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/snapshots", pool)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "GetSnapshotsEntries",
		"addr":    addr,
		"section": "rest",
	})

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetAllSnapshotsRCode)
	if err != nil {
		s.l.Warnf("Unable to get snapshot list for pool %s", pool)
		return nil, err
	}

	var snaps []ResourceSnapshotShort
	entries := ResultEntries{Entries: &snaps}
	rsp := GeneralResponse{Data: &entries}
	//var rsp GeneralResponse
	//{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		// each entry is decoded and checked for its required fields
		checked, results, errD := decodeEntries[ResourceSnapshotShort](body)
		if errD != nil {
			return nil, errD
		}
		return &ResultEntries{Results: results, Entries: &checked}, nil
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}

func (s *RestEndpoint) GetNASVolumesEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes", pool)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "GetNASVolumesEntries",
		"addr":    addr,
		"section": "rest",
	})

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, CodeOK)
	if err != nil {
		s.l.Warnf("Unable to get list of shares")
		return nil, err
	}

	var snaps []ResourceNASVolume
	entries := ResultEntries{Entries: &snaps}
	rsp := GeneralResponse{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		// each entry is decoded and checked for its required fields
		checked, results, errD := decodeEntries[ResourceNASVolume](body)
		if errD != nil {
			return nil, errD
		}
		return &ResultEntries{Results: results, Entries: &checked}, nil
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}

func (s *RestEndpoint) GetNASVolumeSnapshotsEntries(ctx context.Context, pool string, nvds string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s", pool, nvds)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "GetNASVolumeSnapshotsEntries",
		"addr":    addr,
		"section": "rest",
	})

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, CodeOK)
	if err != nil {
		s.l.Warnf("Unable to get list of shares")
		return nil, err
	}

	var snaps []ResourceNASVolumeSnapshot
	entries := ResultEntries{Entries: &snaps}
	rsp := GeneralResponse{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		// each entry is decoded and checked for its required fields
		checked, results, errD := decodeEntries[ResourceNASVolumeSnapshot](body)
		if errD != nil {
			return nil, errD
		}
		return &ResultEntries{Results: results, Entries: &checked}, nil
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}

func (s *RestEndpoint) GetNASSnapshotsEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/snapshots", pool)

	l := jcom.LFC(ctx)

	l = l.WithFields(log.Fields{
		"func":    "GetNASSnapshotsEntries",
		"addr":    addr,
		"section": "rest",
	})

	ctx = jcom.WithLogger(ctx, l)

	addr = pagedcSuffix(addr, &page, &dc)

	l.Debugln("Sending")
	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, CodeOK)
	if err != nil {
		s.l.Warnf("Unable to get NAS snapshot list for pool %s", pool)
		return nil, err
	}

	var snaps []ResourceNASSnapshotShort
	entries := ResultEntries{Entries: &snaps}
	rsp := GeneralResponse{Data: &entries}
	//var rsp GeneralResponse
	//{Data: &entries}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	switch stat {
	case CodeOK, CodeCreated:
		// each entry is decoded and checked for its required fields
		checked, results, errD := decodeEntries[ResourceNASSnapshotShort](body)
		if errD != nil {
			return nil, errD
		}
		return &ResultEntries{Results: results, Entries: &checked}, nil
	default:
		if rsp.Error != nil {
			return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
		}
	}
	return nil, ErrorFromErrorT(ctx, rsp.Error, s.l)
}
