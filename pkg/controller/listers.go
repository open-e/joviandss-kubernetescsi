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

package controller

import (
	//"fmt"
	//"strings"
	"context"
	// "encoding/base64"
	// "fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/container-storage-interface/spec/lib/go/csi"

	//jcom "joviandss-kubernetescsi/pkg/common"
	jdrvr "joviandss-kubernetescsi/pkg/driver"
	jrest "joviandss-kubernetescsi/pkg/rest"
)

func completeListResponseFromSnapshotShort(ctx context.Context, lsr *csi.ListSnapshotsResponse, snaps []jrest.ResourceSnapshotShort) (err error) {

	entries := make([]*csi.ListSnapshotsResponse_Entry, len(snaps))
	lsr.Entries = entries
	for i, s := range snaps {
		ts := timestamppb.New(s.Properties.Creation)

		vd, err := jdrvr.NewVolumeDescFromVDS(s.Volume)
		if err != nil {
			return err
		}
		sd, err := jdrvr.NewSnapshotDescFromSDS(vd, s.Name)
		if err != nil {
			return err
		}

		entries[i] = &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SnapshotId:     sd.CSIID(),
				SourceVolumeId: s.Volume,
				CreationTime:   ts,
				ReadyToUse:	true,
			},
		}
	}

	return nil
}

func completeListResponseFromVolumeSnapshot(ctx context.Context, lsr *csi.ListSnapshotsResponse, snaps []jrest.ResourceSnapshot, ld jdrvr.LunDesc) (err error) {

	entries := make([]*csi.ListSnapshotsResponse_Entry, len(snaps))
	lsr.Entries = entries
	for i, s := range snaps {
		ts := timestamppb.New(s.Creation)

		sd, err := jdrvr.NewSnapshotDescFromSDS(ld, s.Name)
		if err != nil {
			return err
		}

		entries[i] = &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SnapshotId:     sd.CSIID(),
				SourceVolumeId: ld.CSIID(),
				CreationTime:   ts,
				ReadyToUse:	true,
			},
		}
	}

	return nil
}

func completeListResponseFromVolume(ctx context.Context, lsr *csi.ListVolumesResponse, vols []jrest.ResourceVolume) (err error) {

	entries := make([]*csi.ListVolumesResponse_Entry, len(vols))
	lsr.Entries = entries
	for i, v := range vols {

		vd, err := jdrvr.NewVolumeDescFromVDS(v.Name)
		if err != nil {
			return err
		}
		var contentSource *csi.VolumeContentSource
		
		osds := v.OriginSnapshot()
		if len(osds) > 0 {
			if jdrvr.IsSDS(osds) {
				if sd, err := jdrvr.NewSnapshotDescFromSDS(vd,osds); err != nil {
					contentSource = &csi.VolumeContentSource{
						Type: &csi.VolumeContentSource_Snapshot{
							Snapshot: &csi.VolumeContentSource_SnapshotSource{
								SnapshotId: sd.CSIID(),
							},
						},
					}
				}
			} else if jdrvr.IsVDS(osds) {
				if vd, err := jdrvr.NewVolumeDescFromVDS(osds); err == nil {
					contentSource = &csi.VolumeContentSource{
						Type: &csi.VolumeContentSource_Volume{
							Volume: &csi.VolumeContentSource_VolumeSource{
								VolumeId: vd.CSIID(),
							},
						},
					}
				}

			}
		}

		entries[i] = &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				CapacityBytes:  v.GetSize(),
				VolumeId:	vd.CSIID(),
				ContentSource: contentSource,
			},
		}
	}

	return nil
}

