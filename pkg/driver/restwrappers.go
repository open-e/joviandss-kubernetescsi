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
	//"fmt"
	//"strings"
	"context"
	"encoding/base64"
	"fmt"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

func RestVolumeEntryBasedID(entry jrest.ResourceVolume) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestSnapshotEntryBasedID(entry jrest.ResourceSnapshotShort) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestSnapshotShortEntryBasedID(entry jrest.ResourceSnapshotShort) string {
	basedid := fmt.Sprintf("%s_%s", base64.StdEncoding.EncodeToString([]byte(entry.Parent.Name)), base64.StdEncoding.EncodeToString([]byte(entry.Parent.Name)))
	return basedid
}

func RestNASVolumeEntryBasedID(entry jrest.ResourceNASVolume) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestNASVolumeSnapshotEntryBasedID(entry jrest.ResourceNASSnapshotShort) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestNASSnapshotShortEntryBasedID(entry jrest.ResourceNASSnapshotShort) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

// keptSnapshot is a snapshot that the cleanup of intermediate snapshots left
// on a volume, with the clones that were read for it.
type keptSnapshot struct {
	Name   string
	Clones []string
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
