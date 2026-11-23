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
	"context"
	"fmt"
	"strings"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// keptSnapshot is a snapshot that the cleanup of intermediate snapshots left
// on a volume, with the clones that were read for it.
type keptSnapshot struct {
	Name   string
	Clones []string
}

// isIntermediate reports whether a snapshot is one the plugin made to clone
// a volume from a volume: it carries the name of the new volume, so its
// name reads as the name of a volume of the plugin. The first letter is not
// enough: the cleanup deletes an intermediate snapshot that has no clones,
// and a snapshot made by hand can start with "v" too.
func isIntermediate(name string) bool {
	if name == "" {
		return false
	}
	_, err := NewVolumeDescFromVDS(name)
	return err == nil
}

// isCSISnapshot reports whether a snapshot has a name the plugin gives to a
// CSI snapshot.
func isCSISnapshot(name string) bool { return name != "" && IsSDS(name) }

// volumeHasSnapshots is the error of a DeleteVolume on a volume that still
// has snapshots after the cleanup of the intermediate ones. Its code is
// RestErrorResourceBusy, which the controller answers FailedPrecondition, as
// CSI asks for a volume with snapshots. The message names what is in the
// way: the volumes cloned from it, its CSI snapshots, and the snapshots the
// plugin did not make.
func volumeHasSnapshots(volume string, kept []keptSnapshot) jrest.RestError {
	var dvols, dsnaps, ncsi []string
	for _, snap := range kept {
		switch {
		case isCSISnapshot(snap.Name):
			dsnaps = append(dsnaps, snap.Name)
		case isIntermediate(snap.Name):
			dvols = append(dvols, snap.Clones...)
		default:
			ncsi = append(ncsi, snap.Name)
		}
	}
	msg := fmt.Sprintf("Volume %s is dependent upon by", volume)
	if len(dvols) > 0 {
		msg += fmt.Sprintf(" volumes: %s", strings.Join(dvols, ","))
	}
	if len(dsnaps) > 0 {
		msg += fmt.Sprintf(" snapshots: %s", strings.Join(dsnaps, ","))
	}
	if len(ncsi) > 0 {
		msg += fmt.Sprintf(" not CSI related snapshots: %s", strings.Join(ncsi, ","))
	}
	return jrest.GetError(jrest.RestErrorResourceBusy, msg)
}

// deleteSnapshotWithoutClones deletes a snapshot only after the appliance
// has answered that it has no clones, since the appliance deletes a NAS
// snapshot that has a clone, and the clone with it. When the clones cannot
// be read, nothing is deleted and the error is returned.
func deleteSnapshotWithoutClones(ctx context.Context,
	clonesOf func(context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError),
	deleteSnapshot func(context.Context) jrest.RestError,
) (deleted bool, clones []jrest.ResourceVolumeSnapshotClones, err jrest.RestError) {
	if clones, err = clonesOf(ctx); err != nil {
		return false, nil, err
	}
	if len(clones) > 0 {
		return false, clones, nil
	}
	if err = deleteSnapshot(ctx); err != nil {
		return false, nil, err
	}
	return true, nil, nil
}

// clonesNames returns the names of the clones of a clones reply.
func clonesNames(clones []jrest.ResourceVolumeSnapshotClones) (names []string) {
	for _, c := range clones {
		names = append(names, c.Name)
	}
	return names
}
