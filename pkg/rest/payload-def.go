/*
Copyright (c) 2019 Open-E, Inc.
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

///////////////////////////////////////////////////////////////////////////////
/// Error message

// ErrorT error message returned by server
type ErrorT struct {
	Class   string
	Errno   int
	Message string
	Url     string
}

// ErrorData response mask
type ErrorData struct {
	Data  interface{}
	Error ErrorT
}

///////////////////////////////////////////////////////////////////////////////
// /pools

// IOStats data on input ourput statistics
type IOStats struct {
	Read   string `json:"read"`
	Write  string `json:"write"`
	chksum string `json:"chksum"`
}

// Disk structure returned by server
type Disk struct {
	Name    string
	Id      string
	Sn      string
	Model   string
	Path    string
	Health  string
	Size    int64
	Iostats IOStats
	Led     string
	Origin  string
}

// VDevice virtual device structure
type VDevice struct {
	Name    string
	Type    string
	Health  string
	Iostats IOStats
	Disks   []Disk
}

// Enabled flag
type Enabled struct {
	Enabled bool
}

///////////////////////////////////////////////////////////////////////////////
/// Pool

type PoolEnabled struct {
	Enabled bool `enabled:"enabled"`
}

// Pool response structure
type Pool struct {
	Available  string      `json: "available"`
	Status     int         `json:"status"`
	Name       string      `json:"name": "Pool-0"`
	Scan       int         `json:"scan"`
	Encryption PoolEnabled `json:"encryption"`
	Iostats    IOStats     `json:"iostats"`
	Vdevs      []VDevice   `json:"vdevs"`
	Health     string      `json:"health"`
	Operation  string      `json:"operation"`
	ID         string      `json:"id"`
	Size       string      `json:"size"`
}

// GetPoolData response mask
type GetPoolData struct {
	Data  Pool
	Error ErrorT
}

const GetPoolRCode = 200

// PoolShort element of a pool list
type PoolShort struct {
	Name       string
	Status     int
	Health     string
	Scan       int
	Operation  string
	Encryption Enabled
	Iostats    IOStats
	Vdevs      []VDevice
}

// GetPoolsData response mask
type GetPoolsData struct {
	Data  []PoolShort
	Error ErrorT
}

// GetPoolsRCode success response code
const GetPoolsRCode = 200

///////////////////////////////////////////////////////////////////////////////
/// Volume

// Volume response structure
type Volume struct {
	Origin               string `json:"origin"`
	Reference            string `json:"referencce"`
	Primarycache         string `json:"primarycache"`
	Logbias              string `json:"logbias"`
	Creation             string `json:"creation"`
	Sync                 string `json:"sync"`
	IsClone              bool   `json:"is_clone"`
	Dedup                string
	Used                 string
	Full_name            string
	Type                 string
	Written              string
	Usedbyrefreservation string
	Compression          string
	Usedbysnapshots      string
	Copies               string
	Compressratio        string
	Readonly             string
	Mlslabel             string
	Secondarycache       string
	Available            string
	Resource_name        string
	Volblocksize         string
	Refcompressratio     string
	Snapdev              string
	Volsize              string
	Reservation          string
	Usedbychildren       string
	Usedbydataset        string
	Name                 string
	Checksum             string
	Refreservation       string
}

// GetVolumeData data
type GetVolumeData struct {
	Data  Volume
	Error ErrorT
}

// GetVolumeRCode success response code
const GetVolumeRCode = 200

// GetVolumesData data structure
type GetVolumesData struct {
	Data  []Volume `json:"data"`
	Error ErrorT   `json:"error"`
}

// GetVolumesRCode success response code
const GetVolumesRCode = 200

///////////////////////////////////////////////////////////////////////////////
/// Create Volume

// CreateVolume request
type CreateVolume struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// CreateVolumeData data
type CreateVolumeData struct {
	Data CreateVolumeR
}

// CreateVolumeR response
type CreateVolumeR struct {
	Origin    string
	Is_clone  bool
	Full_name string
	Name      string
}

// CreateVolumeRCode success status code
const CreateVolumeRCode = 201

// CreateVolumeECodeExists exit code for volume exists
const CreateVolumeECodeExists = 5

///////////////////////////////////////////////////////////////////////////////
/// Delete volume

// DeleteVolume request
type DeleteVolume struct {
	RecursivelyChildren   bool `json:"recursively_children"`
	RecursivelyDependents bool `json:"recursively_dependents"`
	ForceUmount           bool `json:"force_umount"`
}

// DeleteVolumeData data
type DeleteVolumeData struct {
	Error ErrorT
}

// DeleteVolumeRCode success status code
const DeleteVolumeRCode = 204

///////////////////////////////////////////////////////////////////////////////
/// Create Snapshot

// CreateSnapshot request
type CreateSnapshot struct {
	Snapshot_name string `json:"snapshot_name"`
}

// CreateSnapshotRCode success status code
const CreateSnapshotRCode = 200

// CreateSnapshotECodeExists exit code for snapshot exists
const CreateSnapshotECodeExists = 5

// CreateSnapshotData data
type CreateSnapshotData struct {
	Error ErrorT `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
/// Get Snapshot

// Snapshot structure
type Snapshot struct {
	Referenced       string
	Name             string
	Defer_destroy    string
	Userrefs         string
	Primarycache     string
	Type             string
	Creation         string
	Refcompressratio string
	Compressratio    string
	Written          string
	Used             string
	Clones           string
	Mlslabel         string
	Secondarycache   string
}

// GetSnapshotData data
type GetSnapshotData struct {
	Data  Snapshot
	Error ErrorT
}

// GetSnapshotRCode success status code
const GetSnapshotRCode = 200

// GetSnapshotRCodeDNE exit code for snapshot do not exists
const GetSnapshotRCodeDNE = 500

///////////////////////////////////////////////////////////////////////////////
/// Get Snapshots

// SnapshotProperties structure
type SnapshotProperties struct {
	Creation string
}

// SnapshotShort structure
type SnapshotShort struct {
	Volume     string
	Name       string
	Clones     string
	Properties SnapshotProperties
}

// AllSnapshots structure
type AllSnapshots struct {
	Results int
	Entries []SnapshotShort
}

// GetAllSnapshotsData data
type GetAllSnapshotsData struct {
	Data  AllSnapshots
	Error ErrorT
}

// GetAllSnapshotsRCode success status code
const GetAllSnapshotsRCode = 200

// VolSnapshots structure
type VolSnapshots struct {
	Results int
	Entries []Snapshot
}

// GetVolSnapshotsData data
type GetVolSnapshotsData struct {
	Data  VolSnapshots
	Error ErrorT
}

// GetVolSnapshotsRCode success status code
const GetVolSnapshotsRCode = 200

///////////////////////////////////////////////////////////////////////////////
/// Delete Snapshot

// DeleteSnapshot structure
type DeleteSnapshot struct {
	Recursively_dependents bool
}

// DeleteSnapshotData data
type DeleteSnapshotData struct {
	Error ErrorT
}

// DeleteSnapshotRCode success status code
const DeleteSnapshotRCode = 204

// DeleteSnapshotRCodeBusy snapshot is busy code
const DeleteSnapshotRCodeBusy = 1000

///////////////////////////////////////////////////////////////////////////////
/// Clone volume

// CreateClone request
type CreateClone struct {
	Name     string `json:"name"`
	Snapshot string `json:"snapshot"`
}

// CreateCloneR response
type CreateCloneR struct {
	Origin   string `json:"origin"`
	IsClone  bool   `json:"is_clone"`
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

// CreateCloneData data
type CreateCloneData struct {
	Data  CreateCloneR
	Error ErrorT
}

// CreateCloneRCode success status code
const CreateCloneRCode = 200

///////////////////////////////////////////////////////////////////////////////
/// Delete clone

// DeleteClone request
type DeleteClone struct {
	RecursivelyChildren   bool `json:"recursively_children"`
	RecursivelyDependents bool `json:"recursively_dependents"`
	ForceUmount           bool `json:"force_umount"`
}

// DeleteCloneData data
type DeleteCloneData struct {
	Error ErrorT
}

// DeleteCloneRCode success status code
const DeleteCloneRCode = 204

///////////////////////////////////////////////////////////////////////////////
/// Promote cloned volume

// PromoteClone request
type PromoteClone struct {
	Poolname string `json:"poolname"`
}

// PromoteCloneData response data
type PromoteCloneData struct {
	Error ErrorT
}

// PromoteCloneRCode success status code
const PromoteCloneRCode = 200

///////////////////////////////////////////////////////////////////////////////
// Get Target

type Target struct {
	IncomingUsersActive bool     `json:"incoming_users_active"`
	Name                string   `json:"name"`
	AllowIP             []string `json:"allow_ip"`
	OutgoingUser        string   `json:"outgoing_user"`
	Active              bool     `json:"active"`
	Conflicted          bool     `json:"conflicted"`
	DenyIP              []string `json:"deny_ip"`
}

//GetTargetData data
type GetTargetData struct {
	Data  Target
	Error ErrorT
}

// GetTargetRCode success status code
const GetTargetRCode = 200

// GetTargetRCode success status code
const GetTargetRCodeDoNotExists = 404

///////////////////////////////////////////////////////////////////////////////
/// Create Target

// CreateTarget request
type CreateTarget struct {
	Name                string `json:"name"`
	Active              bool   `json:"active"`
	IncomingUsersActive bool   `json:"incoming_users_active"`
}

// CreateTargetData response data
type CreateTargetData struct {
	Error ErrorT
}

// CreateTargetRCode success status code
const CreateTargetRCode = 201

// DeleteTargetRCode success status code
const DeleteTargetRCode = 204

///////////////////////////////////////////////////////////////////////////////
/// Attach Volume to Target

// AttachToTarget request
type AttachToTarget struct {
	Name string `json:"name"`
	Lun  int    `json:"lun"`
	Mode string `json:"mode"`
}

// AttachToTargetData response data
type AttachToTargetData struct {
	Error ErrorT
}

// AttachToTargetRCode success status code
const AttachToTargetRCode = 201

// DettachFromTargetRCode success status code
const DettachFromTargetRCode = 204

///////////////////////////////////////////////////////////////////////////////
/// Add User to Target

// AddUserToTarget request
type AddUserToTarget struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AddUserToTargetData data structure
type AddUserToTargetData struct {
	Error ErrorT
}

// AddUserToTargetRCode success status code
const AddUserToTargetRCode = 201
