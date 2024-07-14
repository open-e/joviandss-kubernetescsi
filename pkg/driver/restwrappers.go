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
	"encoding/base64"
	"fmt"

	// jcom "joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

func RestVolumeEntryBasedID(entry jrest.ResourceVolume) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestSnapshotEntryBasedID(entry jrest.ResourceSnapshot) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestSnapshotShortEntryBasedID(entry jrest.ResourceSnapshotShort) string {
	basedid := fmt.Sprintf("%s_%s", base64.StdEncoding.EncodeToString([]byte(entry.Volume)), base64.StdEncoding.EncodeToString([]byte(entry.Volume)))
	return basedid
}

func RestNASVolumeEntryBasedID(entry jrest.ResourceNASVolume) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}

func RestNASVolumeSnapshotEntryBasedID(entry jrest.ResourceNASVolumeSnapshot) string {
	basedid := base64.StdEncoding.EncodeToString([]byte(entry.Name))
	return basedid
}
