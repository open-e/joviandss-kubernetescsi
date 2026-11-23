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

// pageQuery adds the paging to the path of a list. Every list request
// carries per_page: without it an appliance returns pages of its own
// default size, and JovianDSS 2 ignores page and returns the first page
// again and again.
func pageQuery(path string, page, perPage int) string {
	return fmt.Sprintf("%s?page=%d&per_page=%d", path, page, perPage)
}

// The calls below return one page of a list, as the appliance sent it: the
// caller counts its entries to find the last page, before it drops any.

// GetVolumeSnapshotsEntries returns one page of the snapshots of one volume.
// On API v4 the entries are short entries, as in the list of all snapshots.
func (s *RestEndpoint) GetVolumeSnapshotsEntries(ctx context.Context, pool string, vname string, page, perPage int) ([]ResourceSnapshotShort, RestError) {
	path := apiPath("/pools/%s/volumes/%s/snapshots", pool, vname).page(page, perPage)
	return readEntries[ResourceSnapshotShort](ctx, s, path, getVolumeSnapshotsEntriesErrors)
}

func (s *RestEndpoint) GetVolumesEntries(ctx context.Context, pool string, page, perPage int) ([]ResourceVolume, RestError) {
	path := apiPath("/pools/%s/volumes", pool).page(page, perPage)
	return readEntries[ResourceVolume](ctx, s, path, getVolumesEntriesErrors)
}

// GetSnapshotsEntries returns one page of the snapshots of the pool. The
// list holds the snapshots of zvols and of NAS volumes together: parent.type
// tells them apart.
func (s *RestEndpoint) GetSnapshotsEntries(ctx context.Context, pool string, page, perPage int) ([]ResourceSnapshotShort, RestError) {
	path := apiPath("/pools/%s/volumes/snapshots", pool).page(page, perPage)
	return readEntries[ResourceSnapshotShort](ctx, s, path, getSnapshotsEntriesErrors)
}

func (s *RestEndpoint) GetNASVolumesEntries(ctx context.Context, pool string, page, perPage int) ([]ResourceNASVolume, RestError) {
	path := apiPath("/pools/%s/nas-volumes", pool).page(page, perPage)
	return readEntries[ResourceNASVolume](ctx, s, path, getNASVolumesEntriesErrors)
}

// GetNASVolumeSnapshotsEntries returns one page of the snapshots of one NAS
// volume. On API v4 the entries are short entries, as in the list of all NAS
// snapshots.
func (s *RestEndpoint) GetNASVolumeSnapshotsEntries(ctx context.Context, pool string, nvds string, page, perPage int) ([]ResourceNASSnapshotShort, RestError) {
	path := apiPath("/pools/%s/nas-volumes/%s/snapshots", pool, nvds).page(page, perPage)
	return readEntries[ResourceNASSnapshotShort](ctx, s, path, getNASVolumeSnapshotsEntriesErrors)
}

// GetNASSnapshotsEntries returns one page of the NAS snapshots of the pool.
// This list too holds the snapshots of both kinds.
func (s *RestEndpoint) GetNASSnapshotsEntries(ctx context.Context, pool string, page, perPage int) ([]ResourceNASSnapshotShort, RestError) {
	path := apiPath("/pools/%s/nas-volumes/snapshots", pool).page(page, perPage)
	return readEntries[ResourceNASSnapshotShort](ctx, s, path, getNASSnapshotsEntriesErrors)
}
