/*
Copyright (c) 2026 Open-E, Inc.
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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// replyFixture is one line of testdata/replies.json: a success reply
// recorded from a JovianDSS 1 appliance on API v4. The fixtures of a call
// are the smallest set of its recorded replies that together shows every
// read field present, every optional read field null or missing where a
// recording shows it so, and every alternative of the reply.
type replyFixture struct {
	Call     string   `json:"call"`
	Type     string   `json:"type"`     // the Go type the reply decodes into
	Shape    string   `json:"shape"`    // object, entries (a page of a list) or items (a plain list)
	File     string   `json:"file"`     // the recorded body, under testdata/
	Recorded string   `json:"recorded"` // the hurl run, file and case it was recorded in
	Shows    []string `json:"shows"`    // what this fixture adds to those before it
	Note     string   `json:"note"`     // what was trimmed or replaced
}

func loadReplyFixtures(t *testing.T) []replyFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "replies.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []replyFixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/replies.json lists no fixture")
	}
	return fixtures
}

func (f replyFixture) body(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(f.File)))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func (f replyFixture) shows(what string) bool {
	for _, s := range f.Shows {
		if s == what {
			return true
		}
	}
	return false
}

// replyDecoder decodes a reply of one type in each of the three shapes and
// returns the number of objects it decoded.
type replyDecoder struct {
	object  func(body []byte) (interface{}, RestError)
	entries func(body []byte) (int, RestError)
	items   func(body []byte) (int, RestError)
}

func decoderOf[T any, PT interface {
	*T
	replyType
}]() replyDecoder {
	return replyDecoder{
		object: func(body []byte) (interface{}, RestError) {
			v := new(T)
			return v, decodeData(body, PT(v))
		},
		entries: func(body []byte) (int, RestError) {
			page, _, err := decodeEntries[T, PT](body)
			return len(page), err
		},
		items: func(body []byte) (int, RestError) {
			items, err := decodeItems[T, PT](body)
			return len(items), err
		},
	}
}

// replyTypes gives for each reply type its decoder, the read fields it
// requires, and those of them that the appliance sends as a string and the
// plugin converts.
var replyTypes = map[string]struct {
	decode    replyDecoder
	required  []string
	converted []string
}{
	"ResourcePool":                 {decoderOf[ResourcePool](), []string{"available"}, []string{"available"}},
	"ResourceVolume":               {decoderOf[ResourceVolume](), []string{"name", "volsize"}, []string{"volsize"}},
	"ResourceSnapshot":             {decoderOf[ResourceSnapshot](), []string{"volsize", "creation"}, []string{"volsize", "creation"}},
	"ResourceSnapshotShort":        {decoderOf[ResourceSnapshotShort](), []string{"name", "parent.name", "properties.creation"}, nil},
	"ResourceVolumeSnapshotClones": {decoderOf[ResourceVolumeSnapshotClones](), []string{"name"}, nil},
	"ResourceTarget":               {decoderOf[ResourceTarget](), []string{"active"}, nil},
	"ResourceNASVolume":            {decoderOf[ResourceNASVolume](), []string{"name"}, nil},
	"ResourceNASVolumeSnapshot":    {decoderOf[ResourceNASVolumeSnapshot](), []string{"name", "creation"}, []string{"creation"}},
	"ResourceNASSnapshotShort":     {decoderOf[ResourceNASSnapshotShort](), []string{"name", "parent.name", "properties.creation"}, nil},
	"ResourceShare":                {decoderOf[ResourceShare](), nil, nil},
}

// decodeFixture decodes a body as the fixture's type and shape, and returns
// the number of objects decoded.
func decodeFixture(t *testing.T, f replyFixture, body []byte) (int, RestError) {
	t.Helper()
	rt, ok := replyTypes[f.Type]
	if !ok {
		t.Fatalf("%s: no decoder for the type %s", f.File, f.Type)
	}
	switch f.Shape {
	case "object":
		_, err := rt.decode.object(body)
		return 1, err
	case "entries":
		return rt.decode.entries(body)
	case "items":
		return rt.decode.items(body)
	}
	t.Fatalf("%s: unknown shape %s", f.File, f.Shape)
	return 0, nil
}

// Every recorded reply decodes with the transcribed types: the check in Go
// that a reply type accepts the replies of its call.
func TestReplyFixturesDecode(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range loadReplyFixtures(t) {
		seen[f.Type] = true
		n, err := decodeFixture(t, f, f.body(t))
		if err != nil {
			t.Errorf("%s: %s", f.File, err.Error())
			continue
		}
		if (f.shows("entries:none") || f.shows("items:none")) && n != 0 {
			t.Errorf("%s: %d objects decoded from an empty list", f.File, n)
		}
		if (f.shows("entries:some") || f.shows("items:some")) && n == 0 {
			t.Errorf("%s: nothing decoded from a list with entries", f.File)
		}
	}
	for name := range replyTypes {
		if !seen[name] {
			t.Errorf("the reply type %s has no fixture", name)
		}
	}
}

// firstObject returns the first object a fixture's type decodes in a reply
// body: the data itself, the first entry of a page or the first item of a
// plain list; nil when the list is empty. Numbers stay as they were sent.
func firstObject(t *testing.T, f replyFixture, body []byte) (reply map[string]interface{}, object map[string]interface{}) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&reply); err != nil {
		t.Fatalf("%s: %v", f.File, err)
	}
	var list []interface{}
	switch f.Shape {
	case "object":
		object, _ = reply["data"].(map[string]interface{})
		return reply, object
	case "entries":
		data, _ := reply["data"].(map[string]interface{})
		list, _ = data["entries"].([]interface{})
	case "items":
		list, _ = reply["data"].([]interface{})
	}
	if len(list) == 0 {
		return reply, nil
	}
	object, _ = list[0].(map[string]interface{})
	return reply, object
}

// changed returns the fixture's body with one field of its first object
// removed (remove), or set to value. ok is false when the body has no
// object to change.
func changed(t *testing.T, f replyFixture, field string, remove bool, value interface{}) (body []byte, ok bool) {
	t.Helper()
	reply, object := firstObject(t, f, f.body(t))
	if object == nil {
		return nil, false
	}
	parts := strings.Split(field, ".")
	for _, p := range parts[:len(parts)-1] {
		inner, isObject := object[p].(map[string]interface{})
		if !isObject {
			t.Fatalf("%s: %s is not an object", f.File, p)
		}
		object = inner
	}
	last := parts[len(parts)-1]
	if _, present := object[last]; !present {
		t.Fatalf("%s: the fixture has no field %s", f.File, field)
	}
	if remove {
		delete(object, last)
	} else {
		object[last] = value
	}
	body, err := json.Marshal(reply)
	if err != nil {
		t.Fatal(err)
	}
	return body, true
}

// A required read field that is missing, or null, fails the reply with an
// error that names the field, in a fixed text: it is never read as zero.
func TestRequiredReplyFieldMissing(t *testing.T) {
	checked := 0
	for _, f := range loadReplyFixtures(t) {
		for _, field := range replyTypes[f.Type].required {
			want := "required reply field " + field + " is missing"
			for _, how := range []string{"removed", "null"} {
				body, ok := changed(t, f, field, how == "removed", nil)
				if !ok {
					continue
				}
				checked++
				_, err := decodeFixture(t, f, body)
				if err == nil {
					t.Errorf("%s with %s %s: decoded without an error", f.File, field, how)
					continue
				}
				if err.GetCode() != RestErrorRPM || !strings.Contains(err.Error(), want) {
					t.Errorf("%s with %s %s: error %q with code %d, want %q with code %d",
						f.File, field, how, err.Error(), err.GetCode(), want, RestErrorRPM)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no required field checked")
	}
	t.Logf("%d replies with a required field removed or null", checked)
}

// A number or a date that the appliance sends as a string and that does
// not parse fails the reply with an error that names the field: a pool is
// never read as full, nor a volume as of size 0.
func TestReplyFieldDoesNotConvert(t *testing.T) {
	checked := 0
	for _, f := range loadReplyFixtures(t) {
		for _, field := range replyTypes[f.Type].converted {
			body, ok := changed(t, f, field, false, "12 parsecs")
			if !ok {
				continue
			}
			checked++
			want := "reply field " + field + ` does not convert: "12 parsecs"`
			_, err := decodeFixture(t, f, body)
			if err == nil || err.GetCode() != RestErrorRPM || !strings.Contains(err.Error(), want) {
				t.Errorf("%s with %s that does not parse: error %v, want %q", f.File, field, err, want)
			}
		}
	}
	// the quota of a NAS volume is optional, and converted when it is sent
	for _, f := range loadReplyFixtures(t) {
		if f.Type == "ResourceNASVolume" && f.Shape == "object" && f.shows("quota:set") {
			body, _ := changed(t, f, "quota", false, "12 parsecs")
			checked++
			if _, err := decodeFixture(t, f, body); err == nil || !strings.Contains(err.Error(), "reply field quota does not convert") {
				t.Errorf("%s with a quota that does not parse: error %v", f.File, err)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no converted field checked")
	}
}

// A field with the wrong JSON type fails the whole reply when the plugin
// reads it, and only then: a field the plugin does not read is kept as raw
// JSON, so its type can never make a reply fail.
func TestReplyFieldOfAnotherType(t *testing.T) {
	for _, f := range loadReplyFixtures(t) {
		if f.Call != "GetVolume" {
			continue
		}
		body, _ := changed(t, f, "name", false, 7)
		if _, err := decodeFixture(t, f, body); err == nil || err.GetCode() != RestErrorRPM {
			t.Errorf("%s with a number as name: error %v, want a decode error", f.File, err)
		}
		// encryption is not read: an object, a string or a number decode alike
		for _, v := range []interface{}{"off", 7, map[string]interface{}{"enabled": false}, nil} {
			body, _ := changed(t, f, "encryption", false, v)
			if _, err := decodeFixture(t, f, body); err != nil {
				t.Errorf("%s with encryption %v: %s", f.File, v, err.Error())
			}
		}
	}
}

// A reply that is not what its call expects is an error, never an empty
// result: an empty list would read as "no snapshots" or "no clones".
func TestReplyOfAnotherShape(t *testing.T) {
	bodies := []string{
		``, `not json`, `null`, `{}`, `{"data": null}`, `{"data": null, "error": null}`,
		`{"data": "text"}`, `{"data": 7}`, `{"data": {"results": 0}}`, `{"data": {"results": 0, "entries": null}}`,
	}
	for _, body := range bodies {
		if _, _, err := decodeEntries[ResourceVolume]([]byte(body)); err == nil {
			t.Errorf("a list reply %q decoded without an error", body)
		}
		if _, err := decodeItems[ResourceVolumeSnapshotClones]([]byte(body)); err == nil {
			t.Errorf("a clones reply %q decoded without an error", body)
		}
		var v ResourceVolume
		if err := decodeData([]byte(body), &v); err == nil {
			t.Errorf("a volume reply %q decoded without an error", body)
		}
	}
	// an empty list is a list
	if page, results, err := decodeEntries[ResourceVolume]([]byte(`{"data": {"results": 0, "entries": []}}`)); err != nil || len(page) != 0 || results != 0 {
		t.Errorf("an empty page: %v, %d entries", err, len(page))
	}
	if items, err := decodeItems[ResourceVolumeSnapshotClones]([]byte(`{"data": []}`)); err != nil || len(items) != 0 {
		t.Errorf("an empty clones reply: %v, %d items", err, len(items))
	}
}

// The accessors give what the recorded replies hold.
func TestReplyAccessors(t *testing.T) {
	for _, f := range loadReplyFixtures(t) {
		if f.Shape != "object" {
			continue
		}
		v, err := replyTypes[f.Type].decode.object(f.body(t))
		if err != nil {
			t.Errorf("%s: %s", f.File, err.Error())
			continue
		}
		_, sent := firstObject(t, f, f.body(t))
		switch r := v.(type) {
		case *ResourcePool:
			if r.GetAvailable() <= 0 || sent["available"] != r.Available {
				t.Errorf("%s: available %d from %v", f.File, r.GetAvailable(), sent["available"])
			}
		case *ResourceVolume:
			if r.GetSize() <= 0 || r.Name == "" {
				t.Errorf("%s: name %q, size %d", f.File, r.Name, r.GetSize())
			}
			isClone := sent["origin"] != nil
			if (r.OriginVolume() != "") != isClone || (r.OriginSnapshot() != "") != isClone {
				t.Errorf("%s: origin %v read as volume %q, snapshot %q", f.File, sent["origin"], r.OriginVolume(), r.OriginSnapshot())
			}
		case *ResourceSnapshot:
			if r.GetSize() <= 0 || r.GetCreation().Year() < 2020 {
				t.Errorf("%s: size %d, creation %s", f.File, r.GetSize(), r.GetCreation())
			}
			if _, has := sent["clones"]; has != (len(r.ClonesNames()) > 0) {
				t.Errorf("%s: clones %v read as %v", f.File, sent["clones"], r.ClonesNames())
			}
		case *ResourceNASVolume:
			if (sent["quota"] != nil) != (r.GetSize() > 0) {
				t.Errorf("%s: quota %v read as size %d", f.File, sent["quota"], r.GetSize())
			}
			isClone := sent["origin"] != nil
			if (r.OriginVolume() != "") != isClone || (r.OriginSnapshot() != "") != isClone {
				t.Errorf("%s: origin %v read as volume %q, snapshot %q", f.File, sent["origin"], r.OriginVolume(), r.OriginSnapshot())
			}
		case *ResourceNASVolumeSnapshot:
			if r.Name == "" || r.GetCreation().Year() < 2020 {
				t.Errorf("%s: name %q, creation %s", f.File, r.Name, r.GetCreation())
			}
			if _, has := sent["clones"]; has != (len(r.ClonesNames()) > 0) {
				t.Errorf("%s: clones %v read as %v", f.File, sent["clones"], r.ClonesNames())
			}
		case *ResourceTarget:
			if sent["active"] != r.Active {
				t.Errorf("%s: active %v read as %v", f.File, sent["active"], r.Active)
			}
		}
	}
	// the entries of a snapshot list: the name, the parent and the time
	for _, f := range loadReplyFixtures(t) {
		if f.Type != "ResourceSnapshotShort" || !f.shows("entries:some") {
			continue
		}
		page, results, err := decodeEntries[ResourceSnapshotShort](f.body(t))
		if err != nil {
			t.Errorf("%s: %s", f.File, err.Error())
			continue
		}
		if results < int64(len(page)) {
			t.Errorf("%s: results %d, %d entries", f.File, results, len(page))
		}
		for _, e := range page {
			if e.Name == "" || e.Parent.Name == "" || e.Properties.GetCreation().Year() < 2020 {
				t.Errorf("%s: entry %q of parent %q created %s", f.File, e.Name, e.Parent.Name, e.Properties.GetCreation())
			}
		}
	}
}
