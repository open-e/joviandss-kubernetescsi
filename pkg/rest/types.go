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

import (
	"encoding/json"
	"net"
	"regexp"
	"strconv"
	"time"
)

const resourceNamePattern = `/([\w\-\/]+)`

const originNamePattern = `(?P<pool>[\w\-\.]+)/(?P<volume>[\w\-\.]+)@(?P<snapshot>[\w\-\.]+)`

var (
	resourceNameRegexp = regexp.MustCompile(resourceNamePattern)
	originNameRegexp   = regexp.MustCompile(originNamePattern)
)

type GeneralResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorT     `json:"error,ommitempty"`
}

type ResultEntries struct {
	Results int64       `json:"results"`
	Entries interface{} `json:"entries"`
}

type ResourceVolume struct {
	Name                 string `json:"name,omitempty"`
	Size                 int64  `json:"size,omitempty"`
	Origin               string `json:"origin,omitempty"`
	Relatime             string `json:"relatime,omitempty"`
	Acltype              string `json:"acltype,omitempty"`
	Vscan                string `json:"vscan,omitempty"`
	FullName             string `json:"full_name,omitempty"`
	UserRefs             string `json:"userrefs,omitempty"`
	PrimaryCache         string `json:"primarycache,omitempty"`
	LogBias              string `json:"logbias,omitempty"`
	Creation             string `json:"creation,omitempty"`
	Sync                 string `json:"sync,omitempty"`
	IsClone              bool   `json:"is_clone,omitempty"`
	Dedup                string `json:"dedup,omitempty"`
	ShareNFS             string `json:"sharenfs,omitempty"`
	ReceiveResumeToken   string `json:"receive_resume_token,omitempty"`
	VolSize              int64  `json:"volsize,omitempty"`
	Referenced           string `json:"referenced,omitempty"`
	ShareSMB             string `json:"sharesmb,omitempty"`
	CreateTxg            string `json:"createtxg,omitempty"`
	Reservation          string `json:"reservation,omitempty"`
	SContext             string `json:"scontext,omitempty"`
	MountPoint           string `json:"mountpoint,omitempty"`
	CaseSensitivity      string `json:"casesensitivity,omitempty"`
	GUID                 string `json:"guid,omitempty"`
	UsedByRefReservation string `json:"usedbyrefreservation,omitempty"`
	DNodeSize            string `json:"dnodesize,omitempty"`
	Written              string `json:"written,omitempty"`
	LogicalUsed          string `json:"logicalused,omitempty"`
	CompressRatio        string `json:"compressratio,omitempty"`
	RootContext          string `json:"rootcontext,omitempty"`
	DefaultSCSIID        string `json:"default_scsi_id,omitempty"`
	Type                 string `json:"type,omitempty"`
	Compression          string `json:"compression,omitempty"`
	Snapdir              string `json:"snapdir,omitempty"`
	Overlay              string `json:"overlay,omitempty"`
	Encryption           string `json:"encryption,omitempty"`
	Xattr                string `json:"xattr,omitempty"`
	VolMode              string `json:"volmode,omitempty"`
	Copies               string `json:"copies,omitempty"`
	SnapshotLimit        string `json:"snapshot_limit,omitempty"`
	AclInherit           string `json:"aclinherit,omitempty"`
	DefContext           string `json:"defcontext,omitempty"`
	ReadOnly             string `json:"readonly,omitempty"`
	Version              string `json:"version,omitempty"`
	RecordSize           string `json:"recordsize,omitempty"`
	FilesystemLimit      string `json:"filesystem_limit,omitempty"`
	Mounted              string `json:"mounted,omitempty"`
	MLSLabel             string `json:"mlslabel,omitempty"`
	SecondaryCache       string `json:"secondarycache,omitempty"`
	RefReservation       string `json:"refreservation,omitempty"`
	Available            string `json:"available,omitempty"`
	SanVolumeID          string `json:"san:volume_id,omitempty"`
	EncryptionRoot       string `json:"encryptionroot,omitempty"`
	Exec                 string `json:"exec,omitempty"`
	RefQuota             string `json:"refquota,omitempty"`
	RefCompressRatio     string `json:"refcompressratio,omitempty"`
	Quota                string `json:"quota,omitempty"`
	UTF8Only             string `json:"utf8only,omitempty"`
	KeyLocation          string `json:"keylocation,omitempty"`
	Snapdev              string `json:"snapdev,omitempty"`
	SnapshotCount        string `json:"snapshot_count,omitempty"`
	FSContext            string `json:"fscontext,omitempty"`
	Clones               string `json:"clones,omitempty"`
	CanMount             string `json:"canmount,omitempty"`
	KeyStatus            string `json:"keystatus,omitempty"`
	Atime                string `json:"atime,omitempty"`
	UsedBySnapshots      string `json:"usedbysnapshots,omitempty"`
	Normalization        string `json:"normalization,omitempty"`
	UsedByChildren       string `json:"usedbychildren,omitempty"`
	VolBlockSize         string `json:"volblocksize,omitempty"`
	UsedByDataset        string `json:"usedbydataset,omitempty"`
	ObjSetID             string `json:"objsetid,omitempty"`
	DeferDestroy         string `json:"defer_destroy,omitempty"`
	PBKDF2Iters          string `json:"pbkdf2iters,omitempty"`
	Checksum             string `json:"checksum,omitempty"`
	RedundantMetadata    string `json:"redundant_metadata,omitempty"`
	FilesystemCount      string `json:"filesystem_count,omitempty"`
	Devices              string `json:"devices,omitempty"`
	KeyFormat            string `json:"keyformat,omitempty"`
	SetUID               string `json:"setuid,omitempty"`
	Used                 string `json:"used,omitempty"`
	LogicalReferenced    string `json:"logicalreferenced,omitempty"`
	Context              string `json:"context,omitempty"`
	Zoned                string `json:"zoned,omitempty"`
	NBMAND               string `json:"nbmand,omitempty"`
}

func (m *ResourceVolume) UnmarshalJSON(data []byte) error {
	type Alias ResourceVolume
	aux := &struct {
		VolSize string `json:"volsize,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.VolSize != "" { // Only parse if non-empty
		parsedVolSize, err := strconv.ParseInt(aux.VolSize, 10, 64)
		if err != nil {
			return err
		}
		m.VolSize = parsedVolSize
	}

	return nil
}

func (v *ResourceVolume) GetSize() int64 {
	return v.VolSize
	// if i, err := strconv.ParseInt(v.VolSize, 10, 64); err != nil {
	// 	return 0
	// } else {
	// 	return i
	// }
}

func (v *ResourceVolume) OriginVolume() string {
	if len(v.Origin) > 0 {
		if originNameRegexp.MatchString(v.Origin) {
			match := originNameRegexp.FindStringSubmatch(v.Origin)
			volume := originNameRegexp.SubexpIndex("volume")
			return match[volume]
		}
	}
	return ""
}

func (v *ResourceVolume) OriginSnapshot() string {
	if len(v.Origin) > 0 {
		if originNameRegexp.MatchString(v.Origin) {
			match := originNameRegexp.FindStringSubmatch(v.Origin)
			snapshot := originNameRegexp.SubexpIndex("snapshot")
			return match[snapshot]
		}
	}
	return ""
}

type ResourceSnapshot struct {
	Referenced        string    `json:"referenced,omitempty"`
	UserRefs          string    `json:"userrefs,omitempty"`
	PrimaryCache      string    `json:"primarycache,omitempty"`
	Creation          time.Time `json:"creation,omitempty"`
	VolSize           int64     `json:"volsize,omitempty"`
	CreateTxg         string    `json:"createtxg,omitempty"`
	GUID              string    `json:"guid,omitempty"`
	CompressRatio     string    `json:"compressratio,omitempty"`
	RootContext       string    `json:"rootcontext,omitempty"`
	Encryption        string    `json:"encryption,omitempty"`
	DefContext        string    `json:"defcontext,omitempty"`
	Written           string    `json:"written,omitempty"`
	Type              string    `json:"type,omitempty"`
	SecondaryCache    string    `json:"secondarycache,omitempty"`
	Used              string    `json:"used,omitempty"`
	RefCompressRatio  string    `json:"refcompressratio,omitempty"`
	FSContext         string    `json:"fscontext,omitempty"`
	ObjSetID          string    `json:"objsetid,omitempty"`
	Name              string    `json:"name,omitempty"`
	DeferDestroy      string    `json:"defer_destroy,omitempty"`
	SANVolumeID       string    `json:"san:volume_id,omitempty"`
	MLSLabel          string    `json:"mlslabel,omitempty"`
	LogicalReferenced string    `json:"logicalreferenced,omitempty"`
	Context           string    `json:"context,omitempty"`
	Clones            string    `json:"clones,omitempty"`
}

func (m *ResourceSnapshot) UnmarshalJSON(data []byte) error {
	type Alias ResourceSnapshot
	aux := &struct {
		Creation string `json:"creation,omitempty"`
		VolSize  string `json:"volsize,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m), // Point Alias to ResourceSnapshot to reuse JSON tags
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	const layout = "2006-01-02 15:04:05"
	if aux.Creation != "" { // Only parse if non-empty
		parsedTime, err := time.Parse(layout, aux.Creation)
		if err != nil {
			return err
		}
		m.Creation = parsedTime
	}

	if aux.VolSize != "" { // Only parse if non-empty
		parsedVolSize, err := strconv.ParseInt(aux.VolSize, 10, 64)
		if err != nil {
			return err
		}
		m.VolSize = parsedVolSize
	}

	return nil
}

func (s *ResourceSnapshot) ClonesNames() (clones []string) {
	if len(s.Clones) > 0 {
		matches := resourceNameRegexp.FindAllStringSubmatch(s.Clones, -1)

		for _, v := range matches {
			clones = append(clones, v[1])
		}
	}
	return clones
}

func (s *ResourceSnapshot) GetSize() int64 {
	return s.VolSize
}

type ResourceSnapshotShortProperties struct {
	Creation     time.Time `json:"creation,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
}

func (m *ResourceSnapshotShortProperties) UnmarshalJSON(data []byte) error {
	type Alias ResourceSnapshotShortProperties
	aux := &struct {
		Creation string `json:"creation,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m), // Point Alias to ResourceSnapshot to reuse JSON tags
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.Creation) > 0 { // Only parse if non-empty
		creationTime, err := strconv.ParseInt(aux.Creation, 10, 64)
		if err != nil {
			return nil
		}
		m.Creation = time.Unix(creationTime, 0)
	}

	return nil
}

type ResourceSnapshotShort struct {
	Volume     string
	Name       string
	Properties ResourceSnapshotShortProperties
}

type ResourceVolumeSnapshotClones struct {
	IsClone  string `json:"is_clone,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Name     string `json:"name,omitempty"`
	Origin   string `json:"origin,omitempty"`
}

type ResourcePool struct {
	Available         int64                    `json:"available,omitempty"`
	Status            int                      `json:"status,omitempty"`
	ImportStatus      ResourcePoolImportStatus `json:"import_status,omitempty"`
	Scan              *interface{}             `json:"scan,omitempty"`
	Encryption        ResourcePoolEncryption   `json:"encryption,omitempty"`
	DeduplicationRate string                   `json:"deduplication_rate,omitempty"`
	SysvolUpgrade     string                   `json:"sysvol_upgrade,omitempty"`
	Vdevs             []ResourcePoolVdev       `json:"vdevs,omitempty"`
	ID                string                   `json:"id,omitempty"`
	Health            string                   `json:"health,omitempty"`
	IOStats           ResourcePoolIOStats      `json:"iostats,omitempty"`
	Operation         string                   `json:"operation,omitempty"`
	Size              string                   `json:"size,omitempty"`
	AutoTrim          bool                     `json:"autotrim,omitempty"`
	Name              string                   `json:"name,omitempty"`
}

func (m *ResourcePool) UnmarshalJSON(data []byte) error {
	type Alias ResourcePool
	aux := &struct {
		Available string `json:"available,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m), // Point Alias to ResourceSnapshot to reuse JSON tags
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.Available) > 0 { // Only parse if non-empty
		available, err := strconv.ParseInt(aux.Available, 10, 64)
		if err != nil {
			return nil
		}
		m.Available = available
	}
	return nil
}

type ResourcePoolImportStatus struct {
	ImportSteps      ResourcePoolImportSteps `json:"import_steps,omitempty"`
	ImportSuccessful bool                    `json:"import_successful,omitempty"`
}

type ResourcePoolImportSteps struct {
	SanSetup          bool `json:"san_setup,omitempty"`
	VipSetup          bool `json:"vip_setup,omitempty"`
	ZfsImport         bool `json:"zfs_import,omitempty"`
	MountSystemVolume bool `json:"mount_system_volume,omitempty"`
	NasSetup          bool `json:"nas_setup,omitempty"`
}

type ResourcePoolEncryption struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ResourcePoolVdev struct {
	Name           string              `json:"name,omitempty"`
	IOStats        ResourcePoolIOStats `json:"iostats,omitempty"`
	Disks          []Disk              `json:"disks,omitempty"`
	Health         string              `json:"health,omitempty"`
	VdevReplacings []string            `json:"vdev_replacings,omitempty"`
	VdevSpares     []string            `json:"vdev_spares,omitempty"`
	Type           string              `json:"type,omitempty"`
}

type Disk struct {
	Origin   string               `json:"origin,omitempty"`
	Slot     string               `json:"slot,omitempty"`
	Led      string               `json:"led,omitempty"`
	Name     string               `json:"name,omitempty"`
	IOStats  ResourcePoolIOStats  `json:"iostats,omitempty"`
	Alias    string               `json:"alias,omitempty"`
	Health   string               `json:"health,omitempty"`
	SN       string               `json:"sn,omitempty"`
	TrimData ResourcePoolTrimData `json:"trim_data,omitempty"`
	Path     *string              `json:"path,omitempty"` // Using *string to handle null values
	Model    string               `json:"model,omitempty"`
	ID       string               `json:"id,omitempty"`
	Size     int64                `json:"size,omitempty"`
}

type ResourcePoolTrimData struct {
	Status   string       `json:"status,omitempty"`
	Progress *interface{} `json:"progress,omitempty"`  // Using *interface{} to handle null values
	TrimTime *interface{} `json:"trim_time,omitempty"` // Using *interface{} to handle null values
}

type ResourcePoolIOStats struct {
	Read   string `json:"read,omitempty"`
	Write  string `json:"write,omitempty"`
	Chksum string `json:"chksum,omitempty"`
}

type ResourceTarget struct {
	Name                string                    `json:"name,omitempty"`
	Active              bool                      `json:"active,omitempty"`
	Conflicted          bool                      `json:"conflicted,omitempty"`
	IncomingUsersActive bool                      `json:"incoming_users_active,omitempty"`
	OutgoingUser        *CreateTargetOutgoingUser `json:"outgoing_user,omitempty"`
	AllowIP             []string                  `json:"allow_ip,omitempty"`
	DenyIP              []string                  `json:"deny_ip,omitempty"`
}

type ShareNFS struct {
	NoRootSquash          bool     `json:"no_root_squash"`
	AllowWriteIP          []net.IP `json:"allow_write_ip"`
	AllSquash             bool     `json:"all_squash"`
	Enabled               bool     `json:"enabled"`
	InsecureConnections   bool     `json:"insecure_connections"`
	SynchronousDataRecord bool     `json:"synchronous_data_record"`
	InsecureLockRequests  bool     `json:"insecure_lock_requests"`
	AllowAccessIP         []net.IP `json:"allow_access_ip"`
}

type ShareSMB struct {
	Enabled           bool   `json:"enabled"`
	ReadOnly          bool   `json:"read_only"`
	Visible           bool   `json:"visible"`
	HandlingLargeDirs bool   `json:"handling_large_dirs"`
	DefaultCase       string `json:"default_case"`
	InheritOwner      bool   `json:"inherit_owner"`
	InheritPerms      bool   `json:"inherit_perms"`
	AccessMode        string `json:"access_mode"`
	Timemachine       bool   `json:"timemachine"`
	Recycle           bool   `json:"recycle"`
	RecycleName       string `json:"recycle_name"`
}

type ResourceShare struct {
	Comment    string   `json:"comment"`
	Name       string   `json:"name"`
	RealPath   string   `json:"real_path"`
	NFS        ShareNFS `json:"nfs"`
	SMB        ShareSMB `json:"smb"`
	Active     bool     `json:"active"`
	Path       string   `json:"path"`
	Conflicted bool     `json:"conflicted"`
}

type ResourceNASVolume struct {
	ID                        string       `json:"id,omitempty"`
	Name                      string       `json:"name"`
	FullName                  string       `json:"fullName,omitempty"`
	ScreenName                string       `json:"screenName,omitempty"`
	Deduplication             DedupSetting `json:"deduplication,omitempty"`
	Compression               Compression  `json:"compression,omitempty"`
	RecordSize                int64        `json:"recordsize,omitempty"`
	Sync                      SyncSetting  `json:"sync,omitempty"`
	LogBias                   LogBias      `json:"logBias,omitempty"`
	PrimaryCache              CacheSetting `json:"primaryCache,omitempty"`
	SecondaryCache            CacheSetting `json:"secondaryCache,omitempty"`
	AccessTime                AtimeSetting `json:"accessTime,omitempty"`
	Copies                    Copies       `json:"copies,omitempty"`
	Quota                     int64        `json:"quota,omitempty"`
	Reservation               string       `json:"reservation,omitempty"`
	ResWithDescendents        bool         `json:"resWithDescendents,omitempty"`
	RefquotaWithDescendents   bool         `json:"refquotaWithDescendents,omitempty"`
	Pool                      string       `json:"pool,omitempty"`
	Type                      string       `json:"type,omitempty"`
	SharesCount               int          `json:"sharesCount,omitempty"`
	OdpLocalRetention         bool         `json:"odpLocalRetention,omitempty"`
	OdpLocalEnabled           bool         `json:"odpLocalEnabled,omitempty"`
	OdpRemoteRetention        bool         `json:"odpRemoteRetention,omitempty"`
	OdpRemoteSrcPath          string       `json:"odpRemoteSrcPath,omitempty"`
	OdpRemoteSrcHostname      string       `json:"odpRemoteSrcHostname,omitempty"`
	OdpIsReceivingDataStopped bool         `json:"odpIsReceivingDataStopped,omitempty"`
	UsedByTracker             bool         `json:"usedByTracker,omitempty"`
	Origin                    string       `json:"origin,omitempty"`
}

func (v *ResourceNASVolume) GetSize() int64 {
	return v.Quota
}

func (v *ResourceNASVolume) OriginVolume() string {
	if len(v.Origin) > 0 {
		if originNameRegexp.MatchString(v.Origin) {
			match := originNameRegexp.FindStringSubmatch(v.Origin)
			volume := originNameRegexp.SubexpIndex("volume")
			return match[volume]
		}
	}
	return ""
}

func (v *ResourceNASVolume) OriginSnapshot() string {
	if len(v.Origin) > 0 {
		if originNameRegexp.MatchString(v.Origin) {
			match := originNameRegexp.FindStringSubmatch(v.Origin)
			snapshot := originNameRegexp.SubexpIndex("snapshot")
			return match[snapshot]
		}
	}
	return ""
}

func (m *ResourceNASVolume) UnmarshalJSON(data []byte) error {
	type Alias ResourceNASVolume
	aux := &struct {
		Copies     string `json:"copies,omitempty"`
		RecordSize string `json:"recordsize,omitempty"`
		Quota      string `json:"quota,omitempty"`

		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Copies != "" {
		num, err := strconv.Atoi(aux.Copies)
		if err != nil {
			return err
		}
		m.Copies = Copies(num)
	}

	if aux.RecordSize != "" {

		num, err := strconv.ParseInt(aux.RecordSize, 10, 64)
		if err != nil {
			return err
		}
		m.RecordSize = num
	}

	if aux.Quota != "" {

		num, err := strconv.ParseInt(aux.Quota, 10, 64)
		if err != nil {
			return err
		}
		m.Quota = num
	}
	return nil
}

type ResourceNASVolumeSnapshot struct {
	SetUID            string    `json:"setuid,omitempty"`
	AclType           string    `json:"acltype,omitempty"`
	UserRefs          string    `json:"userrefs,omitempty"`
	PrimaryCache      string    `json:"primarycache,omitempty"`
	Creation          time.Time `json:"creation,omitempty"`
	Referenced        string    `json:"referenced,omitempty"`
	CreateTXG         string    `json:"createtxg,omitempty"`
	CaseSensitivity   string    `json:"casesensitivity,omitempty"`
	GUID              string    `json:"guid,omitempty"`
	XAttr             string    `json:"xattr,omitempty"`
	CompressRatio     string    `json:"compressratio,omitempty"`
	RootContext       string    `json:"rootcontext,omitempty"`
	UTF8Only          string    `json:"utf8only,omitempty"`
	Encryption        string    `json:"encryption,omitempty"`
	DefContext        string    `json:"defcontext,omitempty"`
	Written           string    `json:"written,omitempty"`
	Version           string    `json:"version,omitempty"`
	MlsLabel          string    `json:"mlslabel,omitempty"`
	SecondaryCache    string    `json:"secondarycache,omitempty"`
	Used              string    `json:"used,omitempty"`
	Exec              string    `json:"exec,omitempty"`
	RefCompressRatio  string    `json:"refcompressratio,omitempty"`
	FsContext         string    `json:"fscontext,omitempty"`
	Normalization     string    `json:"normalization,omitempty"`
	ObjSetID          string    `json:"objsetid,omitempty"`
	Name              string    `json:"name"`
	DeferDestroy      string    `json:"defer_destroy,omitempty"`
	Type              string    `json:"type,omitempty"`
	Devices           string    `json:"devices,omitempty"`
	LogicalReferenced string    `json:"logicalreferenced,omitempty"`
	Context           string    `json:"context,omitempty"`
	NbMand            string    `json:"nbmand,omitempty"`
	Clones            string    `json:"clones,omitempty"`
}

func (m *ResourceNASVolumeSnapshot) UnmarshalJSON(data []byte) error {
	type Alias ResourceNASVolumeSnapshot
	aux := &struct {
		Creation string `json:"creation,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	const layout = "2006-01-02 15:04:05"
	if aux.Creation != "" { // Only parse if non-empty
		parsedTime, err := time.Parse(layout, aux.Creation)
		if err != nil {
			return err
		}
		m.Creation = parsedTime
	}

	return nil
}

func (s *ResourceNASVolumeSnapshot) ClonesNames() (clones []string) {
	if len(s.Clones) > 0 {
		matches := resourceNameRegexp.FindAllStringSubmatch(s.Clones, -1)

		for _, v := range matches {
			clones = append(clones, v[1])
		}
	}
	return clones
}

func (s *ResourceNASVolumeSnapshot) GetSize() int64 {
	return 0
}

type ResourceNASSnapshotShortProperties struct {
	Creation     time.Time `json:"creation,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	Guid         string    `json:"guid,omitempty"`
}

func (m *ResourceNASSnapshotShortProperties) UnmarshalJSON(data []byte) error {
	type Alias ResourceNASSnapshotShortProperties
	aux := &struct {
		Creation int64 `json:"creation,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m), // Point Alias to ResourceSnapshot to reuse JSON tags
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.Creation > 0 { // Only parse if non-empty
		//creationTime, err := strconv.ParseInt(aux.Creation, 10, 64)
		//if err != nil {
		//	return nil
		//}
		m.Creation = time.Unix(aux.Creation, 0)
	}

	return nil
}

type ResourceNASSnapshotParent struct {
	Name       string                             `json:"name"`
	Type       JovianDSSStorageProtocolType       `json:"type"`
	Properties ResourceNASSnapshotShortProperties `json:"properties"`
}

func (m *ResourceNASSnapshotParent) UnmarshalJSON(data []byte) error {
	type Alias ResourceNASSnapshotParent
	aux := &struct {
		Type int `json:"copies,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	m.Type = JovianDSSStorageProtocolType(aux.Type)

	return nil
}

type ResourceNASSnapshotShort struct {
	Parent     ResourceNASSnapshotParent          `json:"parent"`
	Name       string                             `json:"name"`
	Properties ResourceNASSnapshotShortProperties `json:"properties"`
}
