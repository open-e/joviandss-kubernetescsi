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

type Primarycache string

const (
	All      Primarycache = "all"
	None     Primarycache = "none"
	Metadata Primarycache = "metadata"
)

// Define Compression as a type with specific allowed values
type Compression string

const (
	CompOff Compression = "off"
	CompOn  Compression = "on"
	LZJB    Compression = "lzjb"
	GZIP    Compression = "gzip"
	GZIP1   Compression = "gzip-1"
	GZIP2   Compression = "gzip-2"
	GZIP3   Compression = "gzip-3"
	GZIP4   Compression = "gzip-4"
	GZIP5   Compression = "gzip-5"
	GZIP6   Compression = "gzip-6"
	GZIP7   Compression = "gzip-7"
	GZIP8   Compression = "gzip-8"
	GZIP9   Compression = "gzip-9"
	ZLE     Compression = "zle"
	LZ4     Compression = "lz4"
)

type SyncSetting string

const (
	Always   SyncSetting = "always"
	Standard SyncSetting = "standard"
	Disabled SyncSetting = "disabled"
)

type CacheSetting string

const (
	CacheAll      CacheSetting = "all"
	CacheNone     CacheSetting = "none"
	CacheMetadata CacheSetting = "metadata"
)

type LogBias string

const (
	LogBiasLatency    LogBias = "latency"
	LogBiasThroughput LogBias = "throughput"
)

type AtimeSetting string

const (
	AtimeOn  AtimeSetting = "on"
	AtimeOff AtimeSetting = "off"
)

type DedupSetting string

const (
	DedupOn           DedupSetting = "on"
	DedupOff          DedupSetting = "off"
	DedupVerify       DedupSetting = "verify"
	DedupSha256       DedupSetting = "sha256"
	DedupSha256Verify DedupSetting = "sha256,verify"
)

// Define Copies as a type
type Copies int

// CreateVolumeProperties struct now includes the new fields
type CreateVolumeProperties struct {
	Primarycache   *Primarycache `json:"primarycache,omitempty"`
	Secondarycache *Primarycache `json:"secondarycache,omitempty"`
	Compression    *Compression  `json:"compression,omitempty"`
	Logbias        *LogBias      `json:"logbias,omitempty"`
	Sync           *SyncSetting  `json:"sync,omitempty"`
	Dedup          *DedupSetting `json:"dedup,omitempty"`
	Copies         *Copies       `json:"copies,omitempty"`
}

type CreateVolumeDescriptor struct {
	Name          string                  `json:"name"`
	Size          string                  `json:"size"`
	Blocksize     *int64                  `json:"blocksize,omitempty"`
	CreateParents *bool                   `json:"create_parents,omitempty"`
	Sparse        *bool                   `json:"sparse,omitempty"`
	Properties    *CreateVolumeProperties `json:"properties,omitempty"`
}

type DeleteVolumeDescriptor struct {
	RecursivelyChildren *bool `json:"recursively_children,omitempty"`
	ForceUmount         *bool `json:"force_umount,omitempty"`
}

type CloneVolumeDescriptor struct {
	Name          string                  `json:"name"`                     // string with the name that will be assigned to clone.
	Snapshot      string                  `json:"snapshot"`                 // string name of the snapshot that clone will be created from.
	CreateParents *bool                   `json:"create_parents,omitempty"` // boolean, if positive creates all non existing parents of dataset where snapshot will be
	Properties    *CreateVolumeProperties `json:"properties,omitempty"`     // object with properties for the new clone.
}

// TODO: Expand spanpshot properties
type CreateSnapshotProperties struct {
	Primarycache   *Primarycache `json:"primarycache,omitempty"`
	Secondarycache *Primarycache `json:"secondarycache,omitempty"`
}

type CreateSnapshotDescriptor struct {
	SnapshotName string                    `json:"snapshot_name"`        // string with name of the new snapshot.
	Recursive    *bool                     `json:"recursive,omitempty"`  // boolean indicating if recursively create snapshots of all descendant datasets
	Properties   *CreateSnapshotProperties `json:"properties,omitempty"` // object containing properties of new snapshot.
}

type DeleteSnapshotDescriptor struct {
	RecursivelyChildren *bool `json:"recursively_children,omitempty"`
	ForceUnmount        *bool `json:"force_umount,omitempty"`
}

type CreateTargetDescriptor struct {
	Name                string                    `json:"name,omitempty"`
	Active              *bool                     `json:"active,omitempty"`
	IncomingUsersActive *bool                     `json:"incoming_users_active,omitempty"`
	OutgoingUser        *CreateTargetOutgoingUser `json:"outgoing_user,omitempty"`
	AllowIP             *[]string                 `json:"allow_ip,omitempty"`
	DenyIP              *[]string                 `json:"deny_ip,omitempty"`
}

type CreateTargetOutgoingUser struct {
	Password *string `json:"password,omitempty"`
	Name     *string `json:"name,omitempty"`
}

type TargetLunDescriptor struct {
	Name      string  `json:"name,omitempty"`
	SCSIID    *string `json:"scsi_id,omitempty"`
	LUN       *int    `json:"lun,omitempty"`
	Mode      *string `json:"mode,omitempty"`
	BlockSize *int    `json:"block_size,omitempty"`
	EUI       *string `json:"eui,omitempty"`
}

type CreateShareDescriptor struct {
	Name    string              `json:"name"`
	Path    string              `json:"path"`
	Comment *string             `json:"comment,omitempty"`
	Active  *bool               `json:"active,omitempty"`
	NFS     *ShareNFSDescriptor `json:"nfs,omitempty"`
	SMB     *ShareSMBDescriptor `json:"smb,omitempty"`
}

type ShareNFSDescriptor struct {
	Enabled               *bool    `json:"enabled,omitempty"`
	AllowAccessIP         []string `json:"allow_access_ip,omitempty"`
	AllowWriteIP          []string `json:"allow_write_ip,omitempty"`
	InsecureConnections   *bool    `json:"insecure_connections,omitempty"`
	SynchronousDataRecord *bool    `json:"synchronous_data_record,omitempty"`
	InsecureLockRequests  *bool    `json:"insecure_lock_requests,omitempty"`
	AllSquash             *bool    `json:"all_squash,omitempty"`
	NoRootSquash          *bool    `json:"no_root_squash,omitempty"`
}

type ShareSMBDescriptor struct {
	Enabled           *bool   `json:"enabled,omitempty"`
	ReadOnly          *bool   `json:"read_only,omitempty"`
	Visible           *bool   `json:"visible,omitempty"`
	HandlingLargeDirs *bool   `json:"handling_large_dirs,omitempty"`
	DefaultCase       *string `json:"default_case,omitempty"`
	InheritOwner      *bool   `json:"inherit_owner,omitempty"`
	InheritPerms      *bool   `json:"inherit_perms,omitempty"`
	AccessMode        *string `json:"access_mode,omitempty"`
	Spotlight         *bool   `json:"spotlight,omitempty"`
	TimeMachine       *bool   `json:"timemachine,omitempty"`
}

type DeleteShareDescriptor struct {
	RecursivelyChildren *bool `json:"recursively_children,omitempty"`
	ForceUnmount        *bool `json:"force_umount,omitempty"`
}

type CreateNASVolumeDescriptor struct {
	Compression             *Compression  `json:"compression,omitempty"`
	PrimaryCache            *CacheSetting `json:"primarycache,omitempty"`
	LogBias                 *LogBias      `json:"logbias,omitempty"`
	Dedup                   *DedupSetting `json:"dedup,omitempty"`
	Copies                  *Copies       `json:"copies,omitempty"`
	Sync                    *SyncSetting  `json:"sync,omitempty"`
	Quota                   *string       `json:"quota,omitempty"`
	RefReservation          *string       `json:"refreservation,omitempty"`
	Reservation             *string       `json:"reservation,omitempty"`
	ResWithDescendents      *bool         `json:"resWithDescendents,omitempty"`
	RefquotaWithDescendents *bool         `json:"refquotaWithDescendents,omitempty"`
	Atime                   *AtimeSetting `json:"atime,omitempty"`
	SecondaryCache          *CacheSetting `json:"secondarycache,omitempty"`
	Name                    string        `json:"name"`
}
