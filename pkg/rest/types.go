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

type Volume struct {
	Name               string `json:"name,omitempty"`
	Size               string `json:"size,omitempty"`
	Origin             string `json:"origin,omitempty"`
	Relatime           string `json:"relatime,omitempty"`
	Acltype            string `json:"acltype,omitempty"`
	Vscan              string `json:"vscan,omitempty"`
	FullName           string `json:"full_name,omitempty"`
	UserRefs           string `json:"userrefs,omitempty"`
	PrimaryCache       string `json:"primarycache,omitempty"`
	LogBias            string `json:"logbias,omitempty"`
	Creation           string `json:"creation,omitempty"`
	Sync               string `json:"sync,omitempty"`
	IsClone            bool   `json:"is_clone,omitempty"`
	Dedup              string `json:"dedup,omitempty"`
	ShareNFS           string `json:"sharenfs,omitempty"`
	ReceiveResumeToken string `json:"receive_resume_token,omitempty"`
	VolSize            string `json:"volsize,omitempty"`
	Referenced         string `json:"referenced,omitempty"`
	ShareSMB           string `json:"sharesmb,omitempty"`
	CreateTxg          string `json:"createtxg,omitempty"`
	Reservation        string `json:"reservation,omitempty"`
	SContext           string `json:"scontext,omitempty"`
	MountPoint         string `json:"mountpoint,omitempty"`
	CaseSensitivity    string `json:"casesensitivity,omitempty"`
	GUID               string `json:"guid,omitempty"`
	UsedByRefReservation string `json:"usedbyrefreservation,omitempty"`
	DNodeSize         string `json:"dnodesize,omitempty"`
	Written           string `json:"written,omitempty"`
	LogicalUsed       string `json:"logicalused,omitempty"`
	CompressRatio     string `json:"compressratio,omitempty"`
	RootContext       string `json:"rootcontext,omitempty"`
	DefaultSCSIID     string `json:"default_scsi_id,omitempty"`
	Type              string `json:"type,omitempty"`
	Compression       string `json:"compression,omitempty"`
	Snapdir           string `json:"snapdir,omitempty"`
	Overlay           string `json:"overlay,omitempty"`
	Encryption        string `json:"encryption,omitempty"`
	Xattr             string `json:"xattr,omitempty"`
	VolMode           string `json:"volmode,omitempty"`
	Copies            string `json:"copies,omitempty"`
	SnapshotLimit     string `json:"snapshot_limit,omitempty"`
	AclInherit        string `json:"aclinherit,omitempty"`
	DefContext        string `json:"defcontext,omitempty"`
	ReadOnly          string `json:"readonly,omitempty"`
	Version           string `json:"version,omitempty"`
	RecordSize        string `json:"recordsize,omitempty"`
	FilesystemLimit   string `json:"filesystem_limit,omitempty"`
	Mounted           string `json:"mounted,omitempty"`
	MLSLabel          string `json:"mlslabel,omitempty"`
	SecondaryCache    string `json:"secondarycache,omitempty"`
	RefReservation    string `json:"refreservation,omitempty"`
	Available         string `json:"available,omitempty"`
	SanVolumeID       string `json:"san:volume_id,omitempty"`
	EncryptionRoot    string `json:"encryptionroot,omitempty"`
	Exec              string `json:"exec,omitempty"`
	RefQuota          string `json:"refquota,omitempty"`
	RefCompressRatio  string `json:"refcompressratio,omitempty"`
	Quota             string `json:"quota,omitempty"`
	UTF8Only          string `json:"utf8only,omitempty"`
	KeyLocation       string `json:"keylocation,omitempty"`
	Snapdev           string `json:"snapdev,omitempty"`
	SnapshotCount     string `json:"snapshot_count,omitempty"`
	FSContext         string `json:"fscontext,omitempty"`
	Clones            string `json:"clones,omitempty"`
	CanMount          string `json:"canmount,omitempty"`
	KeyStatus         string `json:"keystatus,omitempty"`
	Atime             string `json:"atime,omitempty"`
	UsedBySnapshots   string `json:"usedbysnapshots,omitempty"`
	Normalization     string `json:"normalization,omitempty"`
	UsedByChildren    string `json:"usedbychildren,omitempty"`
	VolBlockSize      string `json:"volblocksize,omitempty"`
	UsedByDataset     string `json:"usedbydataset,omitempty"`
	ObjSetID          string `json:"objsetid,omitempty"`
	DeferDestroy      string `json:"defer_destroy,omitempty"`
	PBKDF2Iters       string `json:"pbkdf2iters,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	RedundantMetadata string `json:"redundant_metadata,omitempty"`
	FilesystemCount   string `json:"filesystem_count,omitempty"`
	Devices           string `json:"devices,omitempty"`
	KeyFormat         string `json:"keyformat,omitempty"`
	SetUID            string `json:"setuid,omitempty"`
	Used              string `json:"used,omitempty"`
	LogicalReferenced string `json:"logicalreferenced,omitempty"`
	Context           string `json:"context,omitempty"`
	Zoned             string `json:"zoned,omitempty"`
	NBMAND            string `json:"nbmand,omitempty"`
}