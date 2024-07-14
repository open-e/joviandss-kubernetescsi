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

package driver

import (
	// "crypto/sha256"
	// "fmt"
	// "strings"
	// "time"

	//"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	// jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

type CSIDriver interface {
	// Volume operations
	CreateVolume(ctx context.Context, pool string, vd *VolumeDesc, requiredBytes int64, limitBytes int64) jrest.RestError
	DeleteVolume(ctx context.Context, pool string, vd *VolumeDesc) jrest.RestError
	GetVolume(ctx context.Context, pool string, vd *VolumeDesc) (*jrest.ResourceVolume, jrest.RestError)
	ListAllVolumes(ctx context.Context, pool string, maxret int, token CSIListingToken) (interface{}, *CSIListingToken, jrest.RestError)

	// Snapshot operations
	CreateSnapshot(ctx context.Context, pool string, vd *VolumeDesc, sd *SnapshotDesc) jrest.RestError
	DeleteSnapshot(ctx context.Context, pool string, ld LunDesc, sd *SnapshotDesc) jrest.RestError
	GetSnapshot(ctx context.Context, pool string, vd LunDesc, sd *SnapshotDesc) (*jrest.ResourceSnapshot, jrest.RestError)
	ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeDesc, maxret int, token CSIListingToken) (interface{}, *CSIListingToken, jrest.RestError)
	ListAllSnapshots(ctx context.Context, pool string, maxret int, token CSIListingToken) ([]jrest.ResourceSnapshotShort, *CSIListingToken, jrest.RestError)

	// Volume cloning and restoration
	CreateVolumeFromSnapshot(ctx context.Context, pool string, sd *SnapshotDesc, nvd *VolumeDesc) jrest.RestError
	CreateVolumeFromVolume(ctx context.Context, pool string, vd *VolumeDesc, nvd *VolumeDesc) jrest.RestError

	// Publishing and unpublishing
	PublishVolume(ctx context.Context, pool string, ld LunDesc, publishInfo string, readonly bool) (*map[string]string, jrest.RestError)
	UnpublishVolume(ctx context.Context, pool string, prefix string, ld LunDesc) jrest.RestError

	// Pool operations
	GetPool(ctx context.Context, pool string) (*jrest.ResourcePool, jrest.RestError)

	// Driver-specific operations (these might be different for iSCSI and NFS)
	// GetDriverName() string
	// GetVolumeStats(ctx context.Context, pool string, vd *VolumeDesc) (*VolumeStats, jrest.RestError)
	// ExpandVolume(ctx context.Context, pool string, vd *VolumeDesc, newSize int64) jrest.RestError
}
