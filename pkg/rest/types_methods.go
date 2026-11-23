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

// What is written by hand beside the transcribed reply types of types.go:
// the envelopes of a reply, the decoding of a reply into its type, the check
// of the required read fields, the conversions and the accessors.

import (
	"encoding/json"
	"fmt"
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

// creationLayout is how a snapshot read by name gives its creation time.
const creationLayout = "2006-01-02 15:04:05"

// GeneralResponse is the envelope of every reply: data, or an error.
type GeneralResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorT     `json:"error,ommitempty"`
}

// ResultEntries is the data of a list reply: the length of the whole list
// and the entries of one page.
type ResultEntries struct {
	Results int64       `json:"results"`
	Entries interface{} `json:"entries"`
}

///////////////////////////////////////////////////////////////////////////////
// Decoding

// replyType is a reply type that checks its required read fields against the
// keys of the JSON object it was decoded from, and converts the fields that
// the appliance sends in another form than the plugin reads them.
//
// Go's decoder leaves a field at its zero value when the key is missing or
// null, without an error, so "the appliance sent 0" and "the appliance sent
// nothing" look the same. checkRequired tells them apart for the fields the
// plugin reads and requires: a reply that lacks one fails, by the name of
// the field, and is never read as zero.
type replyType interface {
	checkRequired(keys map[string]json.RawMessage) error
}

// decodeData decodes the data object of a reply into v and checks it.
func decodeData(body []byte, v replyType) RestError {
	var reply struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	return decodeObject(reply.Data, v)
}

// decodeObject decodes one JSON object of a reply into v and checks it: the
// data of a reply, an entry of a list or an item of a clones reply.
func decodeObject(raw json.RawMessage, v replyType) RestError {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		return GetError(RestErrorRPM, "reply data is not an object: "+err.Error())
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	if err := v.checkRequired(keys); err != nil {
		return GetError(RestErrorRPM, err.Error())
	}
	return nil
}

// decodeEntries decodes the page of a list reply, entry by entry, and
// returns it with results, the length of the whole list. A reply without
// entries fails: it is never read as an empty list.
func decodeEntries[T any, PT interface {
	*T
	replyType
}](body []byte) ([]T, int64, RestError) {
	var reply struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, 0, GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(reply.Data, &keys); err != nil {
		return nil, 0, GetError(RestErrorRPM, "reply data is not an object: "+err.Error())
	}
	if err := requireKeys(keys, "entries"); err != nil {
		return nil, 0, GetError(RestErrorRPM, err.Error())
	}
	var data struct {
		Results int64             `json:"results"`
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(reply.Data, &data); err != nil {
		return nil, 0, GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	entries := make([]T, len(data.Entries))
	for i, raw := range data.Entries {
		if err := decodeObject(raw, PT(&entries[i])); err != nil {
			return nil, 0, err
		}
	}
	return entries, data.Results, nil
}

// decodeItems decodes a reply whose data is a plain array, item by item: the
// clones of a snapshot. A reply whose data is not an array fails: it is never
// read as "no clones".
func decodeItems[T any, PT interface {
	*T
	replyType
}](body []byte) ([]T, RestError) {
	var reply struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	var raws []json.RawMessage
	if len(reply.Data) == 0 || reply.Data[0] != '[' {
		return nil, GetError(RestErrorRPM, "reply data is not a list")
	}
	if err := json.Unmarshal(reply.Data, &raws); err != nil {
		return nil, GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	items := make([]T, len(raws))
	for i, raw := range raws {
		if err := decodeObject(raw, PT(&items[i])); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// requireKeys fails on the first of names that the object lacks or holds as
// null. The text is fixed: the live test searches the log for it.
func requireKeys(keys map[string]json.RawMessage, names ...string) error {
	for _, n := range names {
		if raw, ok := keys[n]; !ok || string(raw) == "null" {
			return fmt.Errorf("required reply field %s is missing", n)
		}
	}
	return nil
}

// requireNested does the same for the fields of an object inside the reply,
// such as the parent of a snapshot list entry, and names a missing one
// object.field.
func requireNested(keys map[string]json.RawMessage, object string, names ...string) error {
	if err := requireKeys(keys, object); err != nil {
		return err
	}
	var inner map[string]json.RawMessage
	if err := json.Unmarshal(keys[object], &inner); err != nil {
		return fmt.Errorf("reply field %s does not convert: it is not an object", object)
	}
	for _, n := range names {
		if raw, ok := inner[n]; !ok || string(raw) == "null" {
			return fmt.Errorf("required reply field %s.%s is missing", object, n)
		}
	}
	return nil
}

// parseNumber converts a number that the appliance sends as a string. Its
// error names the field, in a fixed text too, and the value is never read as
// zero.
func parseNumber(field, s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("reply field %s does not convert: %q", field, s)
	}
	return n, nil
}

// parseCreation converts the creation time of a snapshot read by name, a
// date sent as a string.
func parseCreation(field, s string) (time.Time, error) {
	t, err := time.Parse(creationLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("reply field %s does not convert: %q", field, s)
	}
	return t, nil
}

// originPart returns the named part of an origin, pool/volume@snapshot, and
// "" for a resource that is not a clone.
func originPart(origin *string, part string) string {
	if origin == nil || !originNameRegexp.MatchString(*origin) {
		return ""
	}
	match := originNameRegexp.FindStringSubmatch(*origin)
	return match[originNameRegexp.SubexpIndex(part)]
}

// clonesNames returns the names of the clones a snapshot lists, and none for
// a snapshot without clones.
func clonesNames(clones *string) (names []string) {
	if clones == nil {
		return nil
	}
	for _, m := range resourceNameRegexp.FindAllStringSubmatch(*clones, -1) {
		names = append(names, m[1])
	}
	return names
}

///////////////////////////////////////////////////////////////////////////////
// Volumes

func (v *ResourceVolume) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "name", "volsize"); err != nil {
		return err
	}
	v.size, err = parseNumber("volsize", v.VolSize)
	return err
}

// GetSize returns the size of the volume in bytes.
func (v *ResourceVolume) GetSize() int64 {
	return v.size
}

// OriginVolume returns the volume a clone was made from, and "" for a volume
// that is not a clone.
func (v *ResourceVolume) OriginVolume() string {
	return originPart(v.Origin, "volume")
}

// OriginSnapshot returns the snapshot a clone was made from, and "" for a
// volume that is not a clone.
func (v *ResourceVolume) OriginSnapshot() string {
	return originPart(v.Origin, "snapshot")
}

///////////////////////////////////////////////////////////////////////////////
// Snapshots of a volume

func (s *ResourceSnapshot) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "volsize", "creation"); err != nil {
		return err
	}
	if s.size, err = parseNumber("volsize", s.VolSize); err != nil {
		return err
	}
	s.created, err = parseCreation("creation", s.Creation)
	return err
}

// GetSize returns the size of the volume the snapshot was taken of, in bytes.
func (s *ResourceSnapshot) GetSize() int64 {
	return s.size
}

// GetCreation returns the time the snapshot was created.
func (s *ResourceSnapshot) GetCreation() time.Time {
	return s.created
}

// ClonesNames returns the names of the clones of the snapshot, and none for
// a snapshot without clones.
func (s *ResourceSnapshot) ClonesNames() []string {
	return clonesNames(s.Clones)
}

func (s *ResourceSnapshotShort) checkRequired(keys map[string]json.RawMessage) error {
	if err := requireKeys(keys, "name"); err != nil {
		return err
	}
	if err := requireNested(keys, "parent", "name"); err != nil {
		return err
	}
	return requireNested(keys, "properties", "creation")
}

// GetCreation returns the time the snapshot was created.
func (p *ResourceSnapshotShortProperties) GetCreation() time.Time {
	return time.Unix(p.Creation, 0)
}

func (c *ResourceVolumeSnapshotClones) checkRequired(keys map[string]json.RawMessage) error {
	return requireKeys(keys, "name")
}

///////////////////////////////////////////////////////////////////////////////
// Pool

func (p *ResourcePool) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "available"); err != nil {
		return err
	}
	p.available, err = parseNumber("available", p.Available)
	return err
}

// GetAvailable returns the free space of the pool in bytes.
func (p *ResourcePool) GetAvailable() int64 {
	return p.available
}

///////////////////////////////////////////////////////////////////////////////
// Target

func (t *ResourceTarget) checkRequired(keys map[string]json.RawMessage) error {
	return requireKeys(keys, "active")
}

///////////////////////////////////////////////////////////////////////////////
// Share

// Nothing reads a field of a share yet, so none is required.
func (s *ResourceShare) checkRequired(keys map[string]json.RawMessage) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// NAS volumes

func (v *ResourceNASVolume) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "name"); err != nil {
		return err
	}
	// quota is optional: without one the appliance sends null, and the size
	// is 0, as it is today
	v.size = 0
	if v.Quota != nil {
		v.size, err = parseNumber("quota", *v.Quota)
	}
	return err
}

// GetSize returns the quota of the NAS volume in bytes, and 0 for a volume
// without a quota.
func (v *ResourceNASVolume) GetSize() int64 {
	return v.size
}

// OriginVolume returns the NAS volume a clone was made from, and "" for a
// volume that is not a clone.
func (v *ResourceNASVolume) OriginVolume() string {
	return originPart(v.Origin, "volume")
}

// OriginSnapshot returns the snapshot a clone was made from, and "" for a
// volume that is not a clone.
func (v *ResourceNASVolume) OriginSnapshot() string {
	return originPart(v.Origin, "snapshot")
}

///////////////////////////////////////////////////////////////////////////////
// Snapshots of a NAS volume

func (s *ResourceNASVolumeSnapshot) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "name", "creation"); err != nil {
		return err
	}
	s.created, err = parseCreation("creation", s.Creation)
	return err
}

// GetSize returns 0: the reply of a NAS snapshot carries no size.
func (s *ResourceNASVolumeSnapshot) GetSize() int64 {
	return 0
}

// GetCreation returns the time the snapshot was created.
func (s *ResourceNASVolumeSnapshot) GetCreation() time.Time {
	return s.created
}

// ClonesNames returns the names of the clones of the snapshot, and none for
// a snapshot without clones.
func (s *ResourceNASVolumeSnapshot) ClonesNames() []string {
	return clonesNames(s.Clones)
}

func (s *ResourceNASSnapshotShort) checkRequired(keys map[string]json.RawMessage) error {
	if err := requireKeys(keys, "name"); err != nil {
		return err
	}
	if err := requireNested(keys, "parent", "name"); err != nil {
		return err
	}
	return requireNested(keys, "properties", "creation")
}

// GetCreation returns the time the snapshot was created.
func (p *ResourceNASSnapshotShortProperties) GetCreation() time.Time {
	return time.Unix(p.Creation, 0)
}
