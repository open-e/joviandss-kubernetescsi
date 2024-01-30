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
    CompOff     Compression = "off"
    CompOn      Compression = "on"
    LZJB        Compression = "lzjb"
    GZIP        Compression = "gzip"
    GZIP1       Compression = "gzip-1"
    GZIP2	Compression = "gzip-2"
    GZIP3 	Compression = "gzip-3"
    GZIP4 	Compression = "gzip-4"
    GZIP5 	Compression = "gzip-5"
    GZIP6 	Compression = "gzip-6"
    GZIP7 	Compression = "gzip-7"
    GZIP8 	Compression = "gzip-8"
    GZIP9 	Compression = "gzip-9"
    ZLE         Compression = "zle"
    LZ4         Compression = "lz4"
)

// Define Logbias as a type with specific allowed values
type Logbias string

const (
    Latency    Logbias = "latency"
    Throughput Logbias = "throughput"
)

// Define Sync as a type with specific allowed values
type Sync string

const (
    Always   Sync = "always"
    Standard Sync = "standard"
    Disabled Sync = "disabled"
)

// Define Dedup as a type with specific allowed values
type Dedup string

const (
    DedupOff        Dedup = "off"
    DedupOn         Dedup = "on"
    Verify          Dedup = "verify"
    SHA256          Dedup = "sha256"
    SHA256Verify    Dedup = "sha256,verify"
)

// Define Copies as a type
type Copies int

// CreateVolumeProperties struct now includes the new fields
type CreateVolumeProperties struct {
    Primarycache   Primarycache `json:"primarycache,omitempty"`
    Secondarycache Primarycache `json:"secondarycache,omitempty"`
    Compression    Compression  `json:"compression,omitempty"`
    Logbias        Logbias      `json:"logbias,omitempty"`
    Sync           Sync         `json:"sync,omitempty"`
    Dedup          Dedup        `json:"dedup,omitempty"`
    Copies         Copies       `json:"copies,omitempty"`
}

type CreateVolumeDescriptor struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	Blocksize int64 `json:"blocksize,omitempty"`
	CreateParents bool `json:"create_parents,omitempty"`
	Sparse bool `json:"sparse,omitempty"`
	Properties CreateVolumeProperties `json:"properties,omitempty"`
}
