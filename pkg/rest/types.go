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

// The reply types of the JovianDSS REST API, as the plugin decodes them.
//
// They are transcribed from the description of the recorded API in
// jdss-rest-api-spec, commit 434c9cd: the tracked OpenAPI of JovianDSS 1,
// API v4, under empirical/jdss1/v4/openapi3/paths/. The comment on each type
// names the path file its fields were read from. When that repository gains
// recordings, read the types again from the path files that changed since
// that commit, and name the new commit here.
//
// A field the plugin reads has the exact type the recordings show for its
// endpoint, and a comment that marks it required or optional. Every other
// field of a reply is kept as raw JSON, so that its type can never make a
// reply fail to decode. The checks of the required fields, the conversions
// and the accessors are in types_methods.go.

import (
	"encoding/json"
	"time"
)

// ResourceVolume is a zvol or a clone of one: the reply of
// GET pools/<pool>/volumes/<volume>, from the path file
// pools/+poolname+/volumes/+volumename+.yaml, and an entry of
// GET pools/<pool>/volumes, from pools/+poolname+/volumes.yaml. The spec
// keeps a zvol and a clone as two alternatives; a field that one of them
// sends as null is optional here.
type ResourceVolume struct {
	DefaultScsiId string  `json:"default_scsi_id"` // required
	Guid          string  `json:"guid"`            // required: with default_scsi_id and is_clone it tells two entries of one name in a list apart
	IsClone       *bool   `json:"is_clone"`        // optional by decision (ADR 0002, question 65): false on a zvol, true on a clone, in every recording
	Name          string  `json:"name"`            // required
	Origin        *string `json:"origin"`          // optional: null on a zvol, pool/volume@snapshot on a clone
	VolSize       string  `json:"volsize"`         // required: a number sent as a string, read through GetSize

	// Not read: kept as sent, so that no type of theirs can fail.
	Aclinherit           json.RawMessage `json:"aclinherit"`
	Acltype              json.RawMessage `json:"acltype"`
	Atime                json.RawMessage `json:"atime"`
	Available            json.RawMessage `json:"available"`
	Canmount             json.RawMessage `json:"canmount"`
	Casesensitivity      json.RawMessage `json:"casesensitivity"`
	Checksum             json.RawMessage `json:"checksum"`
	Clones               json.RawMessage `json:"clones"`
	Compression          json.RawMessage `json:"compression"`
	Compressratio        json.RawMessage `json:"compressratio"`
	Context              json.RawMessage `json:"context"`
	Copies               json.RawMessage `json:"copies"`
	Createtxg            json.RawMessage `json:"createtxg"`
	Creation             json.RawMessage `json:"creation"`
	Dedup                json.RawMessage `json:"dedup"`
	Defcontext           json.RawMessage `json:"defcontext"`
	DeferDestroy         json.RawMessage `json:"defer_destroy"`
	Devices              json.RawMessage `json:"devices"`
	Dnodesize            json.RawMessage `json:"dnodesize"`
	Encryption           json.RawMessage `json:"encryption"`
	Encryptionroot       json.RawMessage `json:"encryptionroot"`
	Exec                 json.RawMessage `json:"exec"`
	FilesystemCount      json.RawMessage `json:"filesystem_count"`
	FilesystemLimit      json.RawMessage `json:"filesystem_limit"`
	Fscontext            json.RawMessage `json:"fscontext"`
	FullName             json.RawMessage `json:"full_name"`
	Keyformat            json.RawMessage `json:"keyformat"`
	Keylocation          json.RawMessage `json:"keylocation"`
	Keystatus            json.RawMessage `json:"keystatus"`
	Logbias              json.RawMessage `json:"logbias"`
	Logicalreferenced    json.RawMessage `json:"logicalreferenced"`
	Logicalused          json.RawMessage `json:"logicalused"`
	Mlslabel             json.RawMessage `json:"mlslabel"`
	Mounted              json.RawMessage `json:"mounted"`
	Mountpoint           json.RawMessage `json:"mountpoint"`
	Nbmand               json.RawMessage `json:"nbmand"`
	Normalization        json.RawMessage `json:"normalization"`
	Objsetid             json.RawMessage `json:"objsetid"`
	Overlay              json.RawMessage `json:"overlay"`
	Pbkdf2iters          json.RawMessage `json:"pbkdf2iters"`
	Prefetch             json.RawMessage `json:"prefetch"`
	Primarycache         json.RawMessage `json:"primarycache"`
	Quota                json.RawMessage `json:"quota"`
	Readonly             json.RawMessage `json:"readonly"`
	ReceiveResumeToken   json.RawMessage `json:"receive_resume_token"`
	Recordsize           json.RawMessage `json:"recordsize"`
	RedundantMetadata    json.RawMessage `json:"redundant_metadata"`
	Refcompressratio     json.RawMessage `json:"refcompressratio"`
	Referenced           json.RawMessage `json:"referenced"`
	Refquota             json.RawMessage `json:"refquota"`
	Refreservation       json.RawMessage `json:"refreservation"`
	Relatime             json.RawMessage `json:"relatime"`
	Reservation          json.RawMessage `json:"reservation"`
	Rootcontext          json.RawMessage `json:"rootcontext"`
	SanVolumeId          json.RawMessage `json:"san:volume_id"`
	Scontext             json.RawMessage `json:"scontext"`
	Secondarycache       json.RawMessage `json:"secondarycache"`
	Setuid               json.RawMessage `json:"setuid"`
	Sharenfs             json.RawMessage `json:"sharenfs"`
	Sharesmb             json.RawMessage `json:"sharesmb"`
	Snapdev              json.RawMessage `json:"snapdev"`
	Snapdir              json.RawMessage `json:"snapdir"`
	SnapshotCount        json.RawMessage `json:"snapshot_count"`
	SnapshotLimit        json.RawMessage `json:"snapshot_limit"`
	SnapshotsChanged     json.RawMessage `json:"snapshots_changed"`
	SpecialSmallBlocks   json.RawMessage `json:"special_small_blocks"`
	Sync                 json.RawMessage `json:"sync"`
	Type                 json.RawMessage `json:"type"`
	Used                 json.RawMessage `json:"used"`
	Usedbychildren       json.RawMessage `json:"usedbychildren"`
	Usedbydataset        json.RawMessage `json:"usedbydataset"`
	Usedbyrefreservation json.RawMessage `json:"usedbyrefreservation"`
	Usedbysnapshots      json.RawMessage `json:"usedbysnapshots"`
	Userrefs             json.RawMessage `json:"userrefs"`
	Utf8only             json.RawMessage `json:"utf8only"`
	Version              json.RawMessage `json:"version"`
	Volblocksize         json.RawMessage `json:"volblocksize"`
	Volmode              json.RawMessage `json:"volmode"`
	Volthreading         json.RawMessage `json:"volthreading"`
	Vscan                json.RawMessage `json:"vscan"`
	Written              json.RawMessage `json:"written"`
	Xattr                json.RawMessage `json:"xattr"`
	Zoned                json.RawMessage `json:"zoned"`

	size int64 // VolSize as a number, set by checkRequired
}

// ResourceSnapshot is a zvol snapshot read by name: the reply of
// GET pools/<pool>/volumes/<volume>/snapshots/<snapshot>, from the path file
// pools/+poolname+/volumes/+volumename+/snapshots/+snapshotname+.yaml.
// The reply carries no name.
type ResourceSnapshot struct {
	Clones   *string `json:"clones"`   // optional: missing on a snapshot without clones, read through ClonesNames
	Creation string  `json:"creation"` // required: a date sent as a string, read through GetCreation
	VolSize  string  `json:"volsize"`  // required: a number sent as a string, read through GetSize

	// Not read: kept as sent, so that no type of theirs can fail.
	Compressratio     json.RawMessage `json:"compressratio"`
	Context           json.RawMessage `json:"context"`
	Createtxg         json.RawMessage `json:"createtxg"`
	Defcontext        json.RawMessage `json:"defcontext"`
	DeferDestroy      json.RawMessage `json:"defer_destroy"`
	Encryption        json.RawMessage `json:"encryption"`
	Fscontext         json.RawMessage `json:"fscontext"`
	Guid              json.RawMessage `json:"guid"`
	Logicalreferenced json.RawMessage `json:"logicalreferenced"`
	Mlslabel          json.RawMessage `json:"mlslabel"`
	Objsetid          json.RawMessage `json:"objsetid"`
	Prefetch          json.RawMessage `json:"prefetch"`
	Primarycache      json.RawMessage `json:"primarycache"`
	Refcompressratio  json.RawMessage `json:"refcompressratio"`
	Referenced        json.RawMessage `json:"referenced"`
	Rootcontext       json.RawMessage `json:"rootcontext"`
	SanVolumeId       json.RawMessage `json:"san:volume_id"`
	Secondarycache    json.RawMessage `json:"secondarycache"`
	Type              json.RawMessage `json:"type"`
	Used              json.RawMessage `json:"used"`
	Userrefs          json.RawMessage `json:"userrefs"`
	Written           json.RawMessage `json:"written"`

	size    int64     // VolSize as a number, set by checkRequired
	created time.Time // Creation as a time, set by checkRequired
}

// ResourceSnapshotShortProperties is the properties object of a zvol snapshot
// list entry, from pools/+poolname+/volumes/snapshots.yaml.
type ResourceSnapshotShortProperties struct {
	Creation int64  `json:"creation"` // required: seconds since the epoch, read through GetCreation
	Guid     string `json:"guid"`     // required: tells two snapshots of one name in a list apart
}

// ResourceSnapshotParent is the parent object of a zvol snapshot list entry,
// from pools/+poolname+/volumes/snapshots.yaml: the volume the snapshot
// belongs to.
type ResourceSnapshotParent struct {
	Name string `json:"name"` // required
	Type int64  `json:"type"` // required: SnapshotParentVolume for a zvol, SnapshotParentNASVolume for a NAS volume

	// Not read: kept as sent, so that no type of theirs can fail.
	Properties json.RawMessage `json:"properties"`
}

// ResourceSnapshotShort is a short entry of a zvol snapshot list: an entry of
// GET pools/<pool>/volumes/snapshots, from the path file
// pools/+poolname+/volumes/snapshots.yaml.
type ResourceSnapshotShort struct {
	Name       string                          `json:"name"`       // required
	Parent     ResourceSnapshotParent          `json:"parent"`     // required: its name and its type are read
	Properties ResourceSnapshotShortProperties `json:"properties"` // required: its creation and its guid are read
}

// ResourceVolumeSnapshotClones is one clone of a snapshot: an item of
// GET pools/<pool>/volumes/<volume>/snapshots/<snapshot>/clones, from the path
// file pools/+poolname+/volumes/+volumename+/snapshots/+snapshotname+/clones.yaml,
// and of GET pools/<pool>/nas-volumes/<dataset>/snapshots/<snapshot>/clones,
// from pools/+poolname+/nas-volumes/+datasetname+/snapshots/+snapshotname+/clones.yaml,
// whose items carry the name only.
type ResourceVolumeSnapshotClones struct {
	Name string `json:"name"` // required

	// Not read: kept as sent, so that no type of theirs can fail.
	FullName json.RawMessage `json:"full_name"`
	IsClone  json.RawMessage `json:"is_clone"`
}

// ResourcePool is the reply of GET pools/<pool>, from the path file
// pools/+poolname+.yaml.
type ResourcePool struct {
	Available string `json:"available"` // required: a number sent as a string, read through GetAvailable

	// Not read: kept as sent, so that no type of theirs can fail.
	Autoreplace                        json.RawMessage `json:"autoreplace"`
	Autotrim                           json.RawMessage `json:"autotrim"`
	DeduplicationEnabled               json.RawMessage `json:"deduplication_enabled"`
	DeduplicationRate                  json.RawMessage `json:"deduplication_rate"`
	Encryption                         json.RawMessage `json:"encryption"`
	Health                             json.RawMessage `json:"health"`
	Id                                 json.RawMessage `json:"id"`
	ImportStatus                       json.RawMessage `json:"import_status"`
	Iostats                            json.RawMessage `json:"iostats"`
	IsPoolEncryptionPasswordConfigured json.RawMessage `json:"is_pool_encryption_password_configured"`
	LicenseStorageCapacity             json.RawMessage `json:"license_storage_capacity"`
	Name                               json.RawMessage `json:"name"`
	Operation                          json.RawMessage `json:"operation"`
	Scan                               json.RawMessage `json:"scan"`
	Size                               json.RawMessage `json:"size"`
	SpecialSmallBlocks                 json.RawMessage `json:"special_small_blocks"`
	Status                             json.RawMessage `json:"status"`
	SysvolUpgrade                      json.RawMessage `json:"sysvol_upgrade"`
	Vdevs                              json.RawMessage `json:"vdevs"`

	available int64 // Available as a number, set by checkRequired
}

// ResourcePoolImportStatus is the import_status object of a pool, from
// pools/+poolname+.yaml. Nothing reads it.
type ResourcePoolImportStatus struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	ImportSteps      json.RawMessage `json:"import_steps"`
	ImportSuccessful json.RawMessage `json:"import_successful"`
}

// ResourcePoolImportSteps is the import_steps object of a pool's import_status,
// from pools/+poolname+.yaml. Nothing reads it.
type ResourcePoolImportSteps struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	MountSystemVolume json.RawMessage `json:"mount_system_volume"`
	NasSetup          json.RawMessage `json:"nas_setup"`
	NvmetSetup        json.RawMessage `json:"nvmet_setup"`
	SanSetup          json.RawMessage `json:"san_setup"`
	VipSetup          json.RawMessage `json:"vip_setup"`
	ZfsImport         json.RawMessage `json:"zfs_import"`
}

// ResourcePoolEncryption is the encryption object of a pool, from
// pools/+poolname+.yaml. Nothing reads it.
type ResourcePoolEncryption struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	Enabled json.RawMessage `json:"enabled"`
}

// ResourcePoolVdev is an item of the vdevs of a pool, from pools/+poolname+.yaml.
// Nothing reads it.
type ResourcePoolVdev struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	Disks          json.RawMessage `json:"disks"`
	Health         json.RawMessage `json:"health"`
	Iostats        json.RawMessage `json:"iostats"`
	Name           json.RawMessage `json:"name"`
	Type           json.RawMessage `json:"type"`
	VdevReplacings json.RawMessage `json:"vdev_replacings"`
	VdevSpares     json.RawMessage `json:"vdev_spares"`
}

// Disk is an item of the disks of a pool's vdev, from pools/+poolname+.yaml.
// Nothing reads it.
type Disk struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	Alias          json.RawMessage `json:"alias"`
	DiskDevName    json.RawMessage `json:"disk_dev_name"`
	Health         json.RawMessage `json:"health"`
	Id             json.RawMessage `json:"id"`
	InitializeData json.RawMessage `json:"initialize_data"`
	Iostats        json.RawMessage `json:"iostats"`
	Led            json.RawMessage `json:"led"`
	Model          json.RawMessage `json:"model"`
	Name           json.RawMessage `json:"name"`
	Origin         json.RawMessage `json:"origin"`
	PartitionLabel json.RawMessage `json:"partition_label"`
	Path           json.RawMessage `json:"path"`
	Size           json.RawMessage `json:"size"`
	Slot           json.RawMessage `json:"slot"`
	Sn             json.RawMessage `json:"sn"`
	TrimData       json.RawMessage `json:"trim_data"`
	Type           json.RawMessage `json:"type"`
}

// ResourcePoolTrimData is the trim_data object of a disk, from
// pools/+poolname+.yaml. Nothing reads it.
type ResourcePoolTrimData struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	Progress json.RawMessage `json:"progress"`
	Status   json.RawMessage `json:"status"`
	TrimTime json.RawMessage `json:"trim_time"`
}

// ResourcePoolIOStats is the iostats object of a pool, of a vdev and of a disk,
// from pools/+poolname+.yaml. Nothing reads it.
type ResourcePoolIOStats struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	Chksum json.RawMessage `json:"chksum"`
	Read   json.RawMessage `json:"read"`
	Write  json.RawMessage `json:"write"`
}

// ResourceTarget is the reply of GET pools/<pool>/san/iscsi/targets/<target>,
// from the path file pools/+poolname+/san/iscsi/targets/+targetname+.yaml.
type ResourceTarget struct {
	Active bool `json:"active"` // required

	// Not read: kept as sent, so that no type of theirs can fail.
	ActiveFlag          json.RawMessage `json:"active_flag"`
	AllowIp             json.RawMessage `json:"allow_ip"`
	Conflicted          json.RawMessage `json:"conflicted"`
	DenyIp              json.RawMessage `json:"deny_ip"`
	EncryptionLocked    json.RawMessage `json:"encryption_locked"`
	IncomingUsersActive json.RawMessage `json:"incoming_users_active"`
	Name                json.RawMessage `json:"name"`
	OutgoingUser        json.RawMessage `json:"outgoing_user"`
	VipAllowedPortals   json.RawMessage `json:"vip_allowed_portals"`
}

// ShareNFS is the nfs object of a share, from shares/+sharename+.yaml. Nothing
// reads it.
type ShareNFS struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	AllSquash             json.RawMessage `json:"all_squash"`
	AllowAccessIp         json.RawMessage `json:"allow_access_ip"`
	AllowWriteIp          json.RawMessage `json:"allow_write_ip"`
	Enabled               json.RawMessage `json:"enabled"`
	InsecureConnections   json.RawMessage `json:"insecure_connections"`
	InsecureLockRequests  json.RawMessage `json:"insecure_lock_requests"`
	NoRootSquash          json.RawMessage `json:"no_root_squash"`
	SynchronousDataRecord json.RawMessage `json:"synchronous_data_record"`
}

// ShareSMB is the smb object of a share, from shares/+sharename+.yaml. Nothing
// reads it.
type ShareSMB struct {
	// Not read: kept as sent, so that no type of theirs can fail.
	AccessMode        json.RawMessage `json:"access_mode"`
	DefaultCase       json.RawMessage `json:"default_case"`
	Enabled           json.RawMessage `json:"enabled"`
	ExtraFeatures     json.RawMessage `json:"extra_features"`
	Fruit             json.RawMessage `json:"fruit"`
	HandlingLargeDirs json.RawMessage `json:"handling_large_dirs"`
	HideUnreadable    json.RawMessage `json:"hide_unreadable"`
	InheritOwner      json.RawMessage `json:"inherit_owner"`
	InheritPerms      json.RawMessage `json:"inherit_perms"`
	IpRestrictions    json.RawMessage `json:"ip_restrictions"`
	ReadOnly          json.RawMessage `json:"read_only"`
	Recycle           json.RawMessage `json:"recycle"`
	RecycleName       json.RawMessage `json:"recycle_name"`
	Spotlight         json.RawMessage `json:"spotlight"`
	Timemachine       json.RawMessage `json:"timemachine"`
	Visible           json.RawMessage `json:"visible"`
	Xattr             json.RawMessage `json:"xattr"`
}

// ResourceShare is a share: the reply of GET shares/<share>, from the path file
// shares/+sharename+.yaml.
type ResourceShare struct {
	Path string `json:"path"` // required: the path the share was made with, <pool>/<dataset>/<directory>

	// Not read: kept as sent, so that no type of theirs can fail.
	Active               json.RawMessage `json:"active"`
	ActiveFlag           json.RawMessage `json:"active_flag"`
	Comment              json.RawMessage `json:"comment"`
	Conflicted           json.RawMessage `json:"conflicted"`
	EncryptionLocked     json.RawMessage `json:"encryption_locked"`
	MissingConfiguration json.RawMessage `json:"missing_configuration"`
	MissingDirectory     json.RawMessage `json:"missing_directory"`
	Name                 json.RawMessage `json:"name"`
	Nfs                  json.RawMessage `json:"nfs"`
	RealPath             json.RawMessage `json:"real_path"`
	Smb                  json.RawMessage `json:"smb"`
}

// ResourceNASVolume is a NAS volume, a dataset: the reply of
// GET pools/<pool>/nas-volumes/<dataset>, from the path file
// pools/+poolname+/nas-volumes/+datasetname+.yaml, and an entry of
// GET pools/<pool>/nas-volumes, from pools/+poolname+/nas-volumes.yaml.
type ResourceNASVolume struct {
	Name   string  `json:"name"`   // required
	Origin *string `json:"origin"` // optional: null on a volume that is not a clone
	Quota  *string `json:"quota"`  // optional: null without a quota, a number sent as a string with one, read through GetSize

	// Not read: kept as sent, so that no type of theirs can fail.
	Atime              json.RawMessage `json:"atime"`
	Available          json.RawMessage `json:"available"`
	Canmount           json.RawMessage `json:"canmount"`
	Clones             json.RawMessage `json:"clones"`
	Compression        json.RawMessage `json:"compression"`
	Copies             json.RawMessage `json:"copies"`
	Creation           json.RawMessage `json:"creation"`
	Dedup              json.RawMessage `json:"dedup"`
	Encryption         json.RawMessage `json:"encryption"`
	FullName           json.RawMessage `json:"full_name"`
	Logbias            json.RawMessage `json:"logbias"`
	Logicalreferenced  json.RawMessage `json:"logicalreferenced"`
	Logicalused        json.RawMessage `json:"logicalused"`
	Primarycache       json.RawMessage `json:"primarycache"`
	PropertySource     json.RawMessage `json:"property_source"`
	Readonly           json.RawMessage `json:"readonly"`
	Recordsize         json.RawMessage `json:"recordsize"`
	Refquota           json.RawMessage `json:"refquota"`
	Refreservation     json.RawMessage `json:"refreservation"`
	Reservation        json.RawMessage `json:"reservation"`
	Secondarycache     json.RawMessage `json:"secondarycache"`
	SnapshotCount      json.RawMessage `json:"snapshot_count"`
	SpecialSmallBlocks json.RawMessage `json:"special_small_blocks"`
	Sync               json.RawMessage `json:"sync"`
	Used               json.RawMessage `json:"used"`
	Usedbychildren     json.RawMessage `json:"usedbychildren"`
	Usedbydataset      json.RawMessage `json:"usedbydataset"`
	Usedbysnapshots    json.RawMessage `json:"usedbysnapshots"`

	size int64 // Quota as a number, 0 without a quota, set by checkRequired
}

// ResourceNASVolumeSnapshot is a NAS snapshot read by name: the reply of
// GET pools/<pool>/nas-volumes/<dataset>/snapshots/<snapshot>, from the path
// file pools/+poolname+/nas-volumes/+datasetname+/snapshots/+snapshotname+.yaml.
// The reply carries no volsize.
type ResourceNASVolumeSnapshot struct {
	Clones   *string `json:"clones"`   // optional: missing on a snapshot without clones, read through ClonesNames
	Creation string  `json:"creation"` // required: a date sent as a string, read through GetCreation
	Name     string  `json:"name"`     // required

	// Not read: kept as sent, so that no type of theirs can fail.
	Acltype           json.RawMessage `json:"acltype"`
	Casesensitivity   json.RawMessage `json:"casesensitivity"`
	Compressratio     json.RawMessage `json:"compressratio"`
	Context           json.RawMessage `json:"context"`
	Createtxg         json.RawMessage `json:"createtxg"`
	Defcontext        json.RawMessage `json:"defcontext"`
	DeferDestroy      json.RawMessage `json:"defer_destroy"`
	Devices           json.RawMessage `json:"devices"`
	Encryption        json.RawMessage `json:"encryption"`
	Exec              json.RawMessage `json:"exec"`
	Fscontext         json.RawMessage `json:"fscontext"`
	Guid              json.RawMessage `json:"guid"`
	Logicalreferenced json.RawMessage `json:"logicalreferenced"`
	Mlslabel          json.RawMessage `json:"mlslabel"`
	Nbmand            json.RawMessage `json:"nbmand"`
	Normalization     json.RawMessage `json:"normalization"`
	Objsetid          json.RawMessage `json:"objsetid"`
	Prefetch          json.RawMessage `json:"prefetch"`
	Primarycache      json.RawMessage `json:"primarycache"`
	Refcompressratio  json.RawMessage `json:"refcompressratio"`
	Referenced        json.RawMessage `json:"referenced"`
	Rootcontext       json.RawMessage `json:"rootcontext"`
	Secondarycache    json.RawMessage `json:"secondarycache"`
	Setuid            json.RawMessage `json:"setuid"`
	Type              json.RawMessage `json:"type"`
	Used              json.RawMessage `json:"used"`
	Userrefs          json.RawMessage `json:"userrefs"`
	Utf8only          json.RawMessage `json:"utf8only"`
	Version           json.RawMessage `json:"version"`
	Written           json.RawMessage `json:"written"`
	Xattr             json.RawMessage `json:"xattr"`

	created time.Time // Creation as a time, set by checkRequired
}

// ResourceNASSnapshotShortProperties is the properties object of a NAS snapshot
// list entry, from pools/+poolname+/nas-volumes/snapshots.yaml.
type ResourceNASSnapshotShortProperties struct {
	Creation int64  `json:"creation"` // required: seconds since the epoch, read through GetCreation
	Guid     string `json:"guid"`     // required: tells two snapshots of one name in a list apart
}

// ResourceNASSnapshotParent is the parent object of a NAS snapshot list entry,
// from pools/+poolname+/nas-volumes/snapshots.yaml: the NAS volume the
// snapshot belongs to.
type ResourceNASSnapshotParent struct {
	Name string `json:"name"` // required
	Type int64  `json:"type"` // required: SnapshotParentVolume for a zvol, SnapshotParentNASVolume for a NAS volume

	// Not read: kept as sent, so that no type of theirs can fail.
	Properties json.RawMessage `json:"properties"`
}

// ResourceNASSnapshotShort is a short entry of a NAS snapshot list: an entry of
// GET pools/<pool>/nas-volumes/snapshots, from the path file
// pools/+poolname+/nas-volumes/snapshots.yaml.
type ResourceNASSnapshotShort struct {
	Name       string                             `json:"name"`       // required
	Parent     ResourceNASSnapshotParent          `json:"parent"`     // required: its name and its type are read
	Properties ResourceNASSnapshotShortProperties `json:"properties"` // required: its creation and its guid are read
}
