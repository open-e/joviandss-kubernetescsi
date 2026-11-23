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
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// errorFixture is one line of testdata/errors.json: a reply recorded from a
// JovianDSS 1 appliance on API v4 for an error that the scenarios file of
// the call lists, and the two codes the call must give it.
type errorFixture struct {
	Call       string `json:"call"`
	Kind       string `json:"kind"`        // the error model of the spec; "" for a reply without an error body
	Status     int    `json:"status"`      // the HTTP status of the recorded reply
	File       string `json:"file"`        // the recorded body, under testdata/
	Recorded   string `json:"recorded"`    // the hurl run, file and case it was recorded in
	Code       int    `json:"code"`        // what GetCode must return
	CodeName   string `json:"code_name"`   // the name of that code, for the reader
	KindCode   int    `json:"kind_code"`   // what GetKindCode must return
	CodeBefore string `json:"code_before"` // what the call returned before the catalogue
}

func loadErrorFixtures(t *testing.T) []errorFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "errors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []errorFixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/errors.json lists no fixture")
	}
	return fixtures
}

func (f errorFixture) body(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(f.File)))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// logRecorder keeps the log entries of a test.
type logRecorder struct {
	entries []*logrus.Entry
}

func (r *logRecorder) Levels() []logrus.Level { return logrus.AllLevels }

func (r *logRecorder) Fire(e *logrus.Entry) error {
	r.entries = append(r.entries, e)
	return nil
}

// testLogger gives a logger that writes nowhere and a record of what was
// logged through it, at every level.
func testLogger() (*logrus.Entry, *logRecorder) {
	lg := logrus.New()
	lg.SetOutput(discard{})
	lg.SetLevel(logrus.TraceLevel)
	rec := &logRecorder{}
	lg.AddHook(rec)
	return logrus.NewEntry(lg), rec
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

// The errno of an error body is a number, null, or missing: three inputs
// to the matching of a catalogue entry. No recording of the plugin's calls
// shows null yet, so its bodies are written by hand.
func TestErrnoThreeCases(t *testing.T) {
	cases := []struct {
		body string
		want errnoField
		text string
	}{
		{`{"error": {"class": "c", "errno": 5, "message": "m"}}`, errnoIs(5), "5"},
		{`{"error": {"class": "c", "errno": 0, "message": "m"}}`, errnoIs(0), "0"},
		{`{"error": {"class": "c", "errno": null, "message": "m"}}`, errnoNull, "null"},
		{`{"error": {"class": "c", "message": "m"}}`, errnoMissing, "missing"},
	}
	for _, c := range cases {
		var reply struct {
			Error *ErrorT `json:"error"`
		}
		if err := json.Unmarshal([]byte(c.body), &reply); err != nil {
			t.Fatalf("%s: %v", c.body, err)
		}
		if reply.Error.Errno != c.want {
			t.Errorf("%s: errno decoded as %+v, want %+v", c.body, reply.Error.Errno, c.want)
		}
		if got := reply.Error.Errno.String(); got != c.text {
			t.Errorf("%s: errno prints as %q, want %q", c.body, got, c.text)
		}
	}
	if errnoIs(0) == errnoNull || errnoIs(0) == errnoMissing || errnoNull == errnoMissing {
		t.Error("errno 0, null and missing must be three different values")
	}

	// An entry written for one case matches neither of the other two.
	msg := "Volume name v is not attached to target t"
	class := "opene.exceptions.ItemNotFoundError"
	entry := kindItemNotFoundErrorVolumeNameIsNotAttachedTo.means(RestErrorResourceDNELun) // errno missing
	for _, c := range []struct {
		errno errnoField
		want  bool
	}{{errnoMissing, true}, {errnoNull, false}, {errnoIs(0), false}} {
		e := &ErrorT{Class: &class, Errno: c.errno, Message: &msg}
		if got := entry.matches(404, e); got != c.want {
			t.Errorf("entry with errno missing against errno %s: match %v, want %v", c.errno, got, c.want)
		}
	}
}

// Every kind of error has a code of its own inside the range of the layer
// that its class shows, and a pattern that Go's regexp compiled: the spec
// states the patterns for Python and JavaScript.
func TestErrorKinds(t *testing.T) {
	if len(errorKinds) == 0 {
		t.Fatal("no kind of error")
	}
	seen := map[int]bool{}
	for _, k := range errorKinds {
		if seen[k.kind] {
			t.Errorf("kind code %d is used twice", k.kind)
		}
		seen[k.kind] = true
		layer := layerOf(k.class)
		if layer == 0 {
			t.Errorf("kind %d: class %q is of no layer known", k.kind, k.class)
			continue
		}
		if k.kind <= layer || k.kind >= layer+1000 {
			t.Errorf("kind %d with class %q is outside the range of its layer, %d", k.kind, k.class, layer)
		}
	}
	for class, want := range map[string]int{
		"":                                      LayerNoClass,
		"builtins.KeyError":                     LayerPython,
		"zfslib.wrap.zfs.ZfsCmdError":           LayerZfs,
		"opene.exceptions.ItemNotFoundError":    LayerProduct,
		"jsonschema.exceptions.ValidationError": LayerSchema,
		"werkzeug.exceptions.NotFound":          LayerRouting,
		"somewhere.else.Error":                  0,
	} {
		if got := layerOf(class); got != want {
			t.Errorf("layerOf(%q) = %d, want %d", class, got, want)
		}
	}
}

// Every error fixture of a call, passed through the handler with the
// catalogue of that call, gets the two codes its entry gives. The first is
// the code the call returned for that reply before the catalogue, but where
// the design record decides another (testdata/errors.json: code_before).
func TestErrorFixturesThroughTheHandler(t *testing.T) {
	l, _ := testLogger()
	ctx := withTestLogger(context.Background(), l)
	var s RestEndpoint
	calls := map[string]bool{}
	for _, f := range loadErrorFixtures(t) {
		entries, ok := errorCatalogue[f.Call]
		if !ok {
			t.Errorf("%s: no catalogue for the call %s", f.File, f.Call)
			continue
		}
		calls[f.Call] = true
		err := s.failedReply(ctx, "GET", "path", entries, f.Status, f.body(t), nil)
		if err == nil {
			t.Errorf("%s: no error", f.File)
			continue
		}
		if err.GetCode() != f.Code {
			t.Errorf("%s: code %d, want %d (%s)", f.File, err.GetCode(), f.Code, f.CodeName)
		}
		if err.GetKindCode() != f.KindCode {
			t.Errorf("%s: kind code %d, want %d", f.File, err.GetKindCode(), f.KindCode)
		}
		if f.Status == 404 && !isNotFound(err.GetCode()) {
			t.Errorf("%s: a 404 got the code %d, which is not a not-found code", f.File, err.GetCode())
		}
	}
	for call := range errorCatalogue {
		if !calls[call] {
			t.Errorf("the call %s has a catalogue and no fixture", call)
		}
	}
	t.Logf("%d error fixtures of %d calls", len(loadErrorFixtures(t)), len(calls))
}

func isNotFound(code int) bool {
	switch code {
	case RestErrorResourceDNE, RestErrorResourceDNEVolume, RestErrorResourceDNETarget,
		RestErrorResourceDNEShare, RestErrorResourceDNEPool, RestErrorResourceDNELun:
		return true
	}
	return false
}

// The same fixtures, served by a fake appliance to the call itself: the
// call ends in the handler with its own catalogue, whatever the status and
// the body.
func TestErrorFixturesThroughTheCalls(t *testing.T) {
	app := newFakeAppliance(t)
	calls := app.calls()
	tested := map[string]bool{}
	for _, f := range loadErrorFixtures(t) {
		call, ok := calls[f.Call]
		if !ok {
			t.Errorf("%s: this package has no call %s", f.File, f.Call)
			continue
		}
		tested[f.Call] = true
		contentType := "application/json"
		if f.Kind == "" {
			contentType = "text/html"
		}
		app.answer(f.Status, contentType, f.body(t))
		err := call()
		if err == nil {
			t.Errorf("%s: the call %s returned no error", f.File, f.Call)
			continue
		}
		if err.GetCode() != f.Code || err.GetKindCode() != f.KindCode {
			t.Errorf("%s: the call %s returned the codes %d and %d, want %d (%s) and %d",
				f.File, f.Call, err.GetCode(), err.GetKindCode(), f.Code, f.CodeName, f.KindCode)
		}
	}
	for name := range calls {
		if !tested[name] {
			t.Errorf("the call %s has no error fixture", name)
		}
	}
	t.Logf("%d calls answered with their error fixtures, %d requests sent", len(tested), len(app.seen()))
}

// A failed reply that no entry of its call holds is Unknown, with its
// fields kept in the error and the code of its layer alone beside; a 404
// is "not found" whatever its body; a reply without an error body gets the
// code its status gives. None is a crash, and none is guessed to be "busy"
// or "exists".
func TestFailedReplyOutsideTheCatalogue(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		code     int
		kind     int
		contains []string
	}{
		{"unknown error body", 500,
			`{"error": {"class": "zfslib.wrap.zfs.ZfsCmdError", "errno": 77, "message": "cannot frobnicate 'Pool-0/v'", "url": "u"}}`,
			RestErrorFailureUnknown, LayerZfs,
			[]string{"status 500", `"zfslib.wrap.zfs.ZfsCmdError"`, "errno 77", "cannot frobnicate", "DELETE", "pools/Pool-0/x"}},
		{"unknown class", 500,
			`{"error": {"class": "somewhere.else.Error", "errno": 1, "message": "m"}}`,
			RestErrorFailureUnknown, 0, []string{"somewhere.else.Error"}},
		{"busy is not guessed", 500,
			`{"error": {"class": "opene.storage.zfs.zfs.ZfsOeError", "message": "In order to delete a widget, you must delete all of its clones first."}}`,
			RestErrorFailureUnknown, LayerProduct, []string{"errno missing"}},
		{"errno null is not errno missing", 404,
			`{"error": {"class": "opene.exceptions.ItemNotFoundError", "errno": null, "message": "Target t not exists."}}`,
			RestErrorResourceDNE, LayerProduct, []string{"errno null"}},
		{"404 with an unknown JSON body", 404,
			`{"error": {"class": "werkzeug.exceptions.NotFound", "message": "something new"}}`,
			RestErrorResourceDNE, LayerRouting, []string{"something new"}},
		{"404 without a JSON body", 404, `404 not found`, RestErrorResourceDNE, LayerNoClass, []string{"404 not found"}},
		{"401 without a JSON body", 401, `401 unauthorized`, RestErrorAuthFailed, LayerNoClass, []string{"authentication failed", "pools/Pool-0/x"}},
		{"403 without a JSON body", 403, `forbidden`, RestErrorAuthFailed, LayerNoClass, []string{"authentication failed"}},
		{"500 without a body", 500, ``, RestErrorFailureUnknown, LayerNoClass, []string{"status 500"}},
		{"500 with a body that has no error object", 500, `{"data": null}`, RestErrorFailureUnknown, LayerNoClass, []string{"status 500"}},
		{"500 with error null", 500, `{"data": null, "error": null}`, RestErrorFailureUnknown, LayerNoClass, []string{"status 500"}},
		{"204 where data was expected", 204, ``, RestErrorFailureUnknown, LayerNoClass, []string{"status 204"}},
		{"202 without a body", 202, ``, RestErrorFailureUnknown, LayerNoClass, []string{"status 202"}},
	}
	for _, c := range cases {
		l, rec := testLogger()
		ctx := withTestLogger(context.Background(), l)
		var s RestEndpoint
		err := s.failedReply(ctx, "DELETE", "pools/Pool-0/x", dettachVolumeFromTargetErrors, c.status, []byte(c.body), nil)
		if err == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		if err.GetCode() != c.code || err.GetKindCode() != c.kind {
			t.Errorf("%s: codes %d and %d, want %d and %d", c.name, err.GetCode(), err.GetKindCode(), c.code, c.kind)
		}
		for _, part := range c.contains {
			if !strings.Contains(err.Error(), part) {
				t.Errorf("%s: the error %q does not hold %q", c.name, err.Error(), part)
			}
		}
		// logged once, at Error: the catalogue does not hold the reply
		if len(rec.entries) != 1 || rec.entries[0].Level != logrus.ErrorLevel {
			t.Errorf("%s: %d log lines, want one at Error", c.name, len(rec.entries))
		}
	}
}

// One log line per failed call, in one format: the method, the path, the
// status, the class, the errno, the message and the two codes. An error
// the catalogue knows is logged at Debug; one whose entry says Unknown on
// purpose at Error.
func TestFailedReplyLogsOneLine(t *testing.T) {
	known := `{"error": {"class": "opene.exceptions.ItemNotFoundError", "message": "Target iqn.2026:t not exists.", "url": "u"}}`
	l, rec := testLogger()
	ctx := withTestLogger(context.Background(), l)
	var s RestEndpoint
	err := s.failedReply(ctx, "DELETE", "pools/Pool-0/san/iscsi/targets/t/luns/v", dettachVolumeFromTargetErrors, 404, []byte(known), nil)
	if err.GetCode() != RestErrorResourceDNETarget || err.GetKindCode() != KindItemNotFoundErrorTargetNotExists {
		t.Fatalf("codes %d and %d", err.GetCode(), err.GetKindCode())
	}
	if len(rec.entries) != 1 {
		t.Fatalf("%d log lines, want one", len(rec.entries))
	}
	e := rec.entries[0]
	if e.Level != logrus.DebugLevel {
		t.Errorf("an error of the catalogue is logged at %s, want Debug", e.Level)
	}
	want := map[string]interface{}{
		"method": "DELETE", "path": "pools/Pool-0/san/iscsi/targets/t/luns/v", "status": 404,
		"class": "opene.exceptions.ItemNotFoundError", "errno": "missing", "message": "Target iqn.2026:t not exists.",
		"code": RestErrorResourceDNETarget, "kind": KindItemNotFoundErrorTargetNotExists,
	}
	for k, v := range want {
		if e.Data[k] != v {
			t.Errorf("log field %s is %v, want %v", k, e.Data[k], v)
		}
	}

	// Unknown on purpose: the catalogue of the target create holds the
	// reply, and says the plugin cannot tell what it means.
	onPurpose := `{"error": {"class": "opene.san.target.base.iscsi.target.TargetNameConflictError", "message": "Target already exists in SCST."}}`
	l, rec = testLogger()
	ctx = withTestLogger(context.Background(), l)
	err = s.failedReply(ctx, "POST", "pools/Pool-0/san/iscsi/targets", createTargetErrors, 409, []byte(onPurpose), nil)
	if err.GetCode() != RestErrorFailureUnknown || err.GetKindCode() != KindTargetNameConflictErrorTargetAlreadyExistsInScst {
		t.Fatalf("codes %d and %d", err.GetCode(), err.GetKindCode())
	}
	if len(rec.entries) != 1 || rec.entries[0].Level != logrus.ErrorLevel {
		t.Errorf("an entry that says Unknown on purpose: %d log lines, want one at Error", len(rec.entries))
	}

	// A failure that is not an error of the appliance keeps its code, has
	// 0 as its second code and is logged once, at Error.
	l, rec = testLogger()
	ctx = withTestLogger(context.Background(), l)
	sent := GetError(RestErrorUnableToConnect, "connection refused")
	err = s.failedReply(ctx, "GET", "pools/Pool-0", getPoolErrors, 0, nil, sent)
	if err != sent || err.GetCode() != RestErrorUnableToConnect || err.GetKindCode() != 0 {
		t.Errorf("a failure without a reply came back as %v, codes %d and %d", err, err.GetCode(), err.GetKindCode())
	}
	if len(rec.entries) != 1 || rec.entries[0].Level != logrus.ErrorLevel {
		t.Errorf("a failure without a reply: %d log lines, want one at Error", len(rec.entries))
	}
}

// The message of an error of the appliance is kept on one line.
func TestErrorMessageOnOneLine(t *testing.T) {
	body := `{"error": {"class": "zfslib.wrap.zfs.ZfsCmdError", "errno": 1000, "message": "cannot destroy 'Pool-0/v': volume has children\nuse '-r' to destroy the following datasets:\nPool-0/v@s1\nPool-0/v@s2"}}`
	l, _ := testLogger()
	ctx := withTestLogger(context.Background(), l)
	var s RestEndpoint
	err := s.failedReply(ctx, "DELETE", "pools/Pool-0/volumes/v", deleteVolumeErrors, 500, []byte(body), nil)
	if err.GetCode() != RestErrorResourceBusyVolumeHasSnapshots {
		t.Fatalf("code %d", err.GetCode())
	}
	if strings.Contains(err.Error(), "\n") || !strings.Contains(err.Error(), "Pool-0/v@s1 Pool-0/v@s2") {
		t.Errorf("message not on one line: %q", err.Error())
	}
}
