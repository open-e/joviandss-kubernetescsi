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

// GetVolumeSnapshotsEntries returns one page of the snapshots of one volume.
// On API v4 the entries are short entries, as in the list of all snapshots.
func (s *RestEndpoint) GetVolumeSnapshotsEntries(ctx context.Context, pool string, vname string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots", pool, vname), &page, &dc)
	return readEntries[ResourceSnapshotShort](ctx, s, path, getVolumeSnapshotsEntriesErrors)
}

func (s *RestEndpoint) GetVolumesEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/volumes", pool), &page, &dc)
	return readEntries[ResourceVolume](ctx, s, path, getVolumesEntriesErrors)
}

func (s *RestEndpoint) GetSnapshotsEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/volumes/snapshots", pool), &page, &dc)
	return readEntries[ResourceSnapshotShort](ctx, s, path, getSnapshotsEntriesErrors)
}

func (s *RestEndpoint) GetNASVolumesEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes", pool), &page, &dc)
	return readEntries[ResourceNASVolume](ctx, s, path, getNASVolumesEntriesErrors)
}

// GetNASVolumeSnapshotsEntries returns one page of the snapshots of one NAS
// volume. On API v4 the entries are short entries, as in the list of all NAS
// snapshots.
func (s *RestEndpoint) GetNASVolumeSnapshotsEntries(ctx context.Context, pool string, nvds string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s/snapshots", pool, nvds), &page, &dc)
	return readEntries[ResourceNASSnapshotShort](ctx, s, path, getNASVolumeSnapshotsEntriesErrors)
}

func (s *RestEndpoint) GetNASSnapshotsEntries(ctx context.Context, pool string, page int64, dc int64) (ent *ResultEntries, err RestError) {
	path := pagedcSuffix(fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/snapshots", pool), &page, &dc)
	return readEntries[ResourceNASSnapshotShort](ctx, s, path, getNASSnapshotsEntriesErrors)
}
