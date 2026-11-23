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
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

const (
	testTarget = "iqn.2026-11.com.example:t"
	listQuery  = "?page=3&per_page=100"
)

// callRequests gives for each call the request it sends for the arguments
// of fakeAppliance.calls, and the statuses that are its success.
var callRequests = map[string]struct {
	method string
	path   string
	ok     []int
}{
	"GetPool":                      {"GET", "/api/v4/pools/Pool-0", []int{200}},
	"GetVolume":                    {"GET", "/api/v4/pools/Pool-0/volumes/v", []int{200}},
	"CreateVolume":                 {"POST", "/api/v4/pools/Pool-0/volumes", []int{201, 200}},
	"DeleteVolume":                 {"DELETE", "/api/v4/pools/Pool-0/volumes/v", []int{204, 200}},
	"GetVolumesEntries":            {"GET", "/api/v4/pools/Pool-0/volumes" + listQuery, []int{200}},
	"GetVolumeSnapshot":            {"GET", "/api/v4/pools/Pool-0/volumes/v/snapshots/s", []int{200}},
	"CreateSnapshot":               {"POST", "/api/v4/pools/Pool-0/volumes/v/snapshots", []int{200, 201}},
	"DeleteSnapshot":               {"DELETE", "/api/v4/pools/Pool-0/volumes/v/snapshots/s", []int{204}},
	"GetSnapshotsEntries":          {"GET", "/api/v4/pools/Pool-0/volumes/snapshots" + listQuery, []int{200}},
	"GetVolumeSnapshotsEntries":    {"GET", "/api/v4/pools/Pool-0/volumes/v/snapshots" + listQuery, []int{200}},
	"GetVolumeSnapshotClones":      {"GET", "/api/v4/pools/Pool-0/volumes/v/snapshots/s/clones", []int{200}},
	"CreateClone":                  {"POST", "/api/v4/pools/Pool-0/volumes/v/clone", []int{200, 201}},
	"DeleteClone":                  {"DELETE", "/api/v4/pools/Pool-0/volumes/v/snapshots/s/clones/c", []int{204}},
	"GetTarget":                    {"GET", "/api/v4/pools/Pool-0/san/iscsi/targets/" + testTarget, []int{200}},
	"CreateTarget":                 {"POST", "/api/v4/pools/Pool-0/san/iscsi/targets", []int{201, 200}},
	"DeleteTarget":                 {"DELETE", "/api/v4/pools/Pool-0/san/iscsi/targets/" + testTarget, []int{204}},
	"AttachVolumeToTarget":         {"POST", "/api/v4/pools/Pool-0/san/iscsi/targets/" + testTarget + "/luns", []int{201, 200}},
	"DettachVolumeFromTarget":      {"DELETE", "/api/v4/pools/Pool-0/san/iscsi/targets/" + testTarget + "/luns/v", []int{204}},
	"GetNASVolume":                 {"GET", "/api/v4/pools/Pool-0/nas-volumes/d", []int{200}},
	"CreateNASVolume":              {"POST", "/api/v4/pools/Pool-0/nas-volumes", []int{201, 200}},
	"DeleteNASVolume":              {"DELETE", "/api/v4/pools/Pool-0/nas-volumes/d", []int{204}},
	"GetNASVolumesEntries":         {"GET", "/api/v4/pools/Pool-0/nas-volumes" + listQuery, []int{200}},
	"GetNASVolumeSnapshotsEntries": {"GET", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots" + listQuery, []int{200}},
	"GetNASSnapshotsEntries":       {"GET", "/api/v4/pools/Pool-0/nas-volumes/snapshots" + listQuery, []int{200}},
	"CreateNASSnapshot":            {"POST", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots", []int{201, 200}},
	"DeleteNASSnapshot":            {"DELETE", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots/s", []int{204}},
	"GetNASVolumeSnapshot":         {"GET", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots/s", []int{200}},
	"GetNASVolumeSnapshotClones":   {"GET", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots/s/clones", []int{200}},
	"CreateNASClone":               {"POST", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots/s/clones", []int{200, 201}},
	"DeleteNASClone":               {"DELETE", "/api/v4/pools/Pool-0/nas-volumes/d/snapshots/s/clones/c", []int{204}},
	"CreateShare":                  {"POST", "/api/v4/shares", []int{201, 200}},
	"GetShare":                     {"GET", "/api/v4/shares/d", []int{200}},
}

// successBody gives a body that is a success of the call: the first
// recorded reply of a call that decodes its reply, and the empty reply of
// the appliance for the others.
func successBody(t *testing.T, call string) []byte {
	t.Helper()
	for _, f := range loadReplyFixtures(t) {
		if f.Call == call {
			return f.body(t)
		}
	}
	return []byte(`{"data": null, "error": null}`)
}

// Every call sends the method and the path of API v4 that the recordings
// show for it, with Basic authentication and a JSON body, and takes each
// status of its success as one.
func TestCallsSendTheirRequests(t *testing.T) {
	app := newFakeAppliance(t)
	calls := app.calls()
	for name := range calls {
		if _, ok := callRequests[name]; !ok {
			t.Errorf("the call %s has no expected request", name)
		}
	}
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:"+testRestPassword))
	for name, want := range callRequests {
		call, ok := calls[name]
		if !ok {
			t.Errorf("no call %s", name)
			continue
		}
		for _, status := range want.ok {
			body := successBody(t, name)
			if status == 204 {
				body = nil
			}
			app.answer(status, "application/json", body)
			app.seen()
			if err := call(); err != nil {
				t.Errorf("%s answered %d: %s", name, status, err.Error())
			}
			reqs := app.seen()
			if len(reqs) != 1 {
				t.Errorf("%s sent %d requests, want 1", name, len(reqs))
				continue
			}
			r := reqs[0]
			if r.Method != want.method || r.Path != want.path {
				t.Errorf("%s sent %s %s, want %s %s", name, r.Method, r.Path, want.method, want.path)
			}
			if got := r.Header.Get("Authorization"); got != basic {
				t.Errorf("%s: the request does not carry the Basic authentication of the endpoint", name)
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("%s: Content-Type %q", name, got)
			}
			if len(r.Body) == 0 || r.Body[0] != '{' {
				t.Errorf("%s: the request body is %q, want a JSON object", name, r.Body)
			}
		}
	}
}

// A status that is not a success of the call is an error that names it,
// whatever the body: a 202 on the target create, and a 204 on a GET, which
// carries no data and is never read as an empty object or an empty list.
func TestStatusThatIsNoSuccess(t *testing.T) {
	app := newFakeAppliance(t)
	calls := app.calls()
	app.answer(202, "application/json", []byte(`{"data": null, "error": null}`))
	if err := calls["CreateTarget"](); err == nil || !strings.Contains(err.Error(), "status 202") {
		t.Errorf("CreateTarget answered 202: error %v, want one naming the status", err)
	}
	for name, want := range callRequests {
		if want.method != "GET" {
			continue
		}
		app.answer(204, "application/json", nil)
		if err := calls[name](); err == nil || !strings.Contains(err.Error(), "status 204") {
			t.Errorf("%s answered 204: error %v, want one naming the status", name, err)
		}
		// a success status with a body the call cannot use is an error too
		app.answer(200, "application/json", []byte(`{"data": null, "error": null}`))
		if err := calls[name](); err == nil || err.GetCode() != RestErrorRPM {
			t.Errorf("%s answered 200 without data: error %v, want a decode error", name, err)
		}
	}
}

// The request log is one Debug line for each reply: the method, the path
// and the status. Neither the body of a request, which can hold a CHAP
// password, nor a header, which holds the REST password, reaches the log
// at any level, whether the call succeeds or fails.
func TestRequestLogHoldsNoSecret(t *testing.T) {
	app := newFakeAppliance(t)
	chap := "chap-password-not-for-the-log"
	user := "chap-user"
	yes := true
	create := func() RestError {
		return app.ep.CreateTarget(app.ctx, "Pool-0", &CreateTargetDescriptor{
			Name: testTarget, Active: &yes,
			OutgoingUser: &CreateTargetOutgoingUser{Name: &user, Password: &chap},
		})
	}
	basic := base64.StdEncoding.EncodeToString([]byte("admin:" + testRestPassword))
	replies := []struct {
		status int
		body   string
	}{
		{201, `{"data": null, "error": null}`},
		{409, `{"error": {"class": "opene.san.target.base.iscsi.target.TargetNameConflictError", "message": "Target with name ` + testTarget + ` is already present on Pool-0."}}`},
		{500, `{"error": {"class": "builtins.KeyError", "message": "'something new'"}}`},
		{401, `401 unauthorized`},
		{500, ``},
	}
	for _, reply := range replies {
		app.log.entries = nil
		app.seen()
		app.answer(reply.status, "application/json", []byte(reply.body))
		err := create()
		if (err == nil) != (reply.status == 201) {
			t.Errorf("status %d: error %v", reply.status, err)
		}
		reqs := app.seen()
		if len(reqs) != 1 || !strings.Contains(string(reqs[0].Body), chap) || !strings.Contains(reqs[0].Header.Get("Authorization"), basic) {
			t.Fatalf("status %d: the request does not hold the passwords this test looks for", reply.status)
		}
		requestLines := 0
		for _, e := range app.log.entries {
			line := e.Message
			for k, v := range e.Data {
				line += fmt.Sprintf(" %s=%v", k, v)
			}
			for _, secret := range []string{chap, testRestPassword, basic, "Authorization", "Basic "} {
				if strings.Contains(line, secret) {
					t.Errorf("status %d: a log line holds %q: %s", reply.status, secret, line)
				}
			}
			if e.Level == logrus.DebugLevel && e.Message == fmt.Sprintf("POST api/v4/pools/Pool-0/san/iscsi/targets: %d", reply.status) {
				requestLines++
			}
		}
		if requestLines != 1 {
			t.Errorf("status %d: %d request log lines, want 1", reply.status, requestLines)
		}
		if err != nil && (strings.Contains(err.Error(), chap) || strings.Contains(err.Error(), testRestPassword)) {
			t.Errorf("status %d: the error holds a password: %s", reply.status, err.Error())
		}
	}
}

// A name put into a path stays one path segment: "/", "?", "#", "%" and a
// space reach the appliance escaped, and can add no segment, cut the path
// short or start a query. The query of a list request is not escaped.
func TestNamesAreEscapedIntoOnePathSegment(t *testing.T) {
	app := newFakeAppliance(t)
	app.answer(500, "application/json", nil)
	name := "a/b?c#d%e f"
	escaped := "a%2Fb%3Fc%23d%25e%20f"
	for _, c := range []struct {
		call func()
		want string
	}{
		{func() { app.ep.GetVolume(app.ctx, "Pool-0", name) }, "GET /api/v4/pools/Pool-0/volumes/" + escaped},
		{func() { app.ep.GetVolume(app.ctx, name, "v") }, "GET /api/v4/pools/" + escaped + "/volumes/v"},
		{func() { app.ep.DeleteSnapshot(app.ctx, "Pool-0", "v", name, DeleteSnapshotDescriptor{}) },
			"DELETE /api/v4/pools/Pool-0/volumes/v/snapshots/" + escaped},
		{func() { app.ep.DeleteNASClone(app.ctx, "Pool-0", "d", "s", name) },
			"DELETE /api/v4/pools/Pool-0/nas-volumes/d/snapshots/s/clones/" + escaped},
		{func() { app.ep.DettachVolumeFromTarget(app.ctx, "Pool-0", name, "v") },
			"DELETE /api/v4/pools/Pool-0/san/iscsi/targets/" + escaped + "/luns/v"},
		{func() { app.ep.GetShare(app.ctx, name) }, "GET /api/v4/shares/" + escaped},
		{func() { app.ep.GetVolumeSnapshotsEntries(app.ctx, "Pool-0", name, 2, 100) },
			"GET /api/v4/pools/Pool-0/volumes/" + escaped + "/snapshots?page=2&per_page=100"},
		{func() { app.ep.GetNASVolumesEntries(app.ctx, name, 0, 100) },
			"GET /api/v4/pools/" + escaped + "/nas-volumes?page=0&per_page=100"},
		// the name of a target keeps its colon and its dots
		{func() { app.ep.DeleteTarget(app.ctx, "Pool-0", "iqn.2026-11.com.example:t") },
			"DELETE /api/v4/pools/Pool-0/san/iscsi/targets/iqn.2026-11.com.example:t"},
	} {
		app.seen()
		c.call()
		reqs := app.seen()
		if len(reqs) != 1 {
			t.Errorf("%s: %d requests", c.want, len(reqs))
			continue
		}
		if got := reqs[0].Method + " " + reqs[0].Path; got != c.want {
			t.Errorf("the request %q, want %q", got, c.want)
		}
	}
}

// Escaping does not change an empty name, "." or "..", which would address
// the collection or the parent. No request is sent for such a name.
func TestPathRefusesAnEmptyNameAndDots(t *testing.T) {
	app := newFakeAppliance(t)
	app.seen()
	for _, bad := range []string{"", ".", ".."} {
		for what, call := range map[string]func() RestError{
			"the pool of a volume read":   func() RestError { _, e := app.ep.GetVolume(app.ctx, bad, "v"); return e },
			"the volume of a volume read": func() RestError { _, e := app.ep.GetVolume(app.ctx, "Pool-0", bad); return e },
			"the volume of a delete":      func() RestError { return app.ep.DeleteVolume(app.ctx, "Pool-0", bad, DeleteVolumeDescriptor{}) },
			"the NAS volume of a delete":  func() RestError { return app.ep.DeleteNASVolume(app.ctx, "Pool-0", bad) },
			"the snapshot of a delete":    func() RestError { return app.ep.DeleteNASSnapshot(app.ctx, "Pool-0", "d", bad) },
			"the clone of a delete": func() RestError {
				return app.ep.DeleteClone(app.ctx, "Pool-0", "v", "s", bad, DeleteVolumeDescriptor{})
			},
			"the target of a delete":        func() RestError { return app.ep.DeleteTarget(app.ctx, "Pool-0", bad) },
			"the volume of a LUN detach":    func() RestError { return app.ep.DettachVolumeFromTarget(app.ctx, "Pool-0", "t", bad) },
			"the pool of a create":          func() RestError { return app.ep.CreateVolume(app.ctx, bad, CreateVolumeDescriptor{Name: "v"}) },
			"the volume of a snapshot list": func() RestError { _, e := app.ep.GetVolumeSnapshotsEntries(app.ctx, "Pool-0", bad, 0, 100); return e },
			"the snapshot of a clones read": func() RestError { _, e := app.ep.GetNASVolumeSnapshotClones(app.ctx, "Pool-0", "d", bad); return e },
			"the share of a share read":     func() RestError { _, e := app.ep.GetShare(app.ctx, bad); return e },
		} {
			err := call()
			if err == nil || err.GetCode() != RestErrorArgumentIncorrect || err.GetKindCode() != 0 {
				t.Errorf("%q as %s: error %v, want the code %d", bad, what, err, RestErrorArgumentIncorrect)
			}
		}
	}
	if n := len(app.seen()); n != 0 {
		t.Errorf("%d requests reached the appliance", n)
	}
	// a name that only starts with a dot, or holds dots, is a name
	for _, name := range []string{"...", ".a", "a.", "a..b"} {
		if p := apiPath("/pools/%s", name); p.err != nil || p.path != "api/v4/pools/"+name {
			t.Errorf("the name %q: path %q, error %v", name, p.path, p.err)
		}
	}
}
