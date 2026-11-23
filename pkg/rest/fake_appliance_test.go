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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// withTestLogger puts the logger of a test into a context, as the plugin
// does for the context of a CSI call.
func withTestLogger(ctx context.Context, l *logrus.Entry) context.Context {
	return jcom.WithLogger(ctx, l)
}

// seenRequest is one request the fake appliance got.
type seenRequest struct {
	Method string
	Path   string // with the query string
	Header http.Header
	Body   []byte
}

// fakeAppliance is an HTTPS server that answers every request with the
// reply a test gives it, and a REST endpoint of the plugin set up to talk
// to it. It keeps the requests it got and what the endpoint logged.
type fakeAppliance struct {
	t   *testing.T
	srv *httptest.Server
	ep  RestEndpoint
	ctx context.Context
	log *logRecorder

	mu       sync.Mutex
	replies  []fakeReply // answered in turn; the last one stays
	requests []seenRequest
}

// fakeReply is one reply of the fake appliance.
type fakeReply struct {
	status      int
	contentType string
	body        []byte
}

func newFakeAppliance(t *testing.T) *fakeAppliance {
	t.Helper()
	app := &fakeAppliance{t: t, replies: []fakeReply{{http.StatusOK, "application/json", nil}}}
	app.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ := io.ReadAll(r.Body)
		app.mu.Lock()
		app.requests = append(app.requests, seenRequest{Method: r.Method, Path: r.URL.RequestURI(), Header: r.Header.Clone(), Body: got})
		reply := app.replies[0]
		if len(app.replies) > 1 {
			app.replies = app.replies[1:]
		}
		app.mu.Unlock()
		w.Header().Set("Content-Type", reply.contentType)
		w.WriteHeader(reply.status)
		w.Write(reply.body)
	}))
	t.Cleanup(app.srv.Close)

	u, err := url.Parse(app.srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	l, rec := testLogger()
	app.log = rec
	cfg := jcom.RestEndpointCfg{
		Addrs:       []string{u.Hostname()},
		Port:        port,
		Prot:        "https",
		User:        "admin",
		Pass:        testRestPassword,
		IdleTimeOut: "30s",
		RetryDelay:  "1ms", // a GET answered 204 is sent again at once
	}
	if err := SetupEndpoint(&app.ep, &cfg, l); err != nil {
		t.Fatal(err)
	}
	app.ctx = withTestLogger(context.Background(), l)
	return app
}

// testRestPassword is the REST password of the fake appliance: a string
// that no log line may hold.
const testRestPassword = "rest-password-not-for-the-log"

// answer sets the reply to every request from now on.
func (app *fakeAppliance) answer(status int, contentType string, body []byte) {
	app.answers(fakeReply{status, contentType, body})
}

// answers sets the replies to the next requests, one each; the last one
// answers every request after them.
func (app *fakeAppliance) answers(replies ...fakeReply) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.replies = replies
}

// seen returns the requests got so far and forgets them.
func (app *fakeAppliance) seen() []seenRequest {
	app.mu.Lock()
	defer app.mu.Unlock()
	out := app.requests
	app.requests = nil
	return out
}

// calls gives every call of the package that sends a request, by its name,
// with arguments that make a well-formed request.
func (app *fakeAppliance) calls() map[string]func() RestError {
	ep, ctx := &app.ep, app.ctx
	yes := true
	return map[string]func() RestError{
		"GetPool":   func() RestError { _, e := ep.GetPool(ctx, "Pool-0"); return e },
		"GetVolume": func() RestError { _, e := ep.GetVolume(ctx, "Pool-0", "v"); return e },
		"CreateVolume": func() RestError {
			return ep.CreateVolume(ctx, "Pool-0", CreateVolumeDescriptor{Name: "v", Size: "1073741824"})
		},
		"DeleteVolume":      func() RestError { return ep.DeleteVolume(ctx, "Pool-0", "v", DeleteVolumeDescriptor{}) },
		"GetVolumesEntries": func() RestError { _, e := ep.GetVolumesEntries(ctx, "Pool-0", 0, 0); return e },
		"GetVolumeSnapshot": func() RestError { _, e := ep.GetVolumeSnapshot(ctx, "Pool-0", "v", "s"); return e },
		"CreateSnapshot": func() RestError {
			return ep.CreateSnapshot(ctx, "Pool-0", "v", &CreateSnapshotDescriptor{SnapshotName: "s"})
		},
		"DeleteSnapshot": func() RestError {
			return ep.DeleteSnapshot(ctx, "Pool-0", "v", "s", DeleteSnapshotDescriptor{})
		},
		"GetSnapshotsEntries": func() RestError { _, e := ep.GetSnapshotsEntries(ctx, "Pool-0", 0, 0); return e },
		"GetVolumeSnapshotsEntries": func() RestError {
			_, e := ep.GetVolumeSnapshotsEntries(ctx, "Pool-0", "v", 0, 0)
			return e
		},
		"GetVolumeSnapshotClones": func() RestError { _, e := ep.GetVolumeSnapshotClones(ctx, "Pool-0", "v", "s"); return e },
		"CreateClone": func() RestError {
			return ep.CreateClone(ctx, "Pool-0", "v", CloneVolumeDescriptor{Name: "c", Snapshot: "s"})
		},
		"DeleteClone": func() RestError {
			return ep.DeleteClone(ctx, "Pool-0", "v", "s", "c", DeleteVolumeDescriptor{})
		},
		"GetTarget": func() RestError { _, e := ep.GetTarget(ctx, "Pool-0", "iqn.2026-11.com.example:t"); return e },
		"CreateTarget": func() RestError {
			return ep.CreateTarget(ctx, "Pool-0", &CreateTargetDescriptor{Name: "iqn.2026-11.com.example:t", Active: &yes})
		},
		"DeleteTarget": func() RestError { return ep.DeleteTarget(ctx, "Pool-0", "iqn.2026-11.com.example:t") },
		"AttachVolumeToTarget": func() RestError {
			return ep.AttachVolumeToTarget(ctx, "Pool-0", "iqn.2026-11.com.example:t", &TargetLunDescriptor{Name: "v"})
		},
		"DettachVolumeFromTarget": func() RestError {
			return ep.DettachVolumeFromTarget(ctx, "Pool-0", "iqn.2026-11.com.example:t", "v")
		},
		"GetNASVolume":         func() RestError { _, e := ep.GetNASVolume(ctx, "Pool-0", "d"); return e },
		"CreateNASVolume":      func() RestError { return ep.CreateNASVolume(ctx, "Pool-0", &CreateNASVolumeDescriptor{Name: "d"}) },
		"DeleteNASVolume":      func() RestError { return ep.DeleteNASVolume(ctx, "Pool-0", "d") },
		"GetNASVolumesEntries": func() RestError { _, e := ep.GetNASVolumesEntries(ctx, "Pool-0", 0, 0); return e },
		"GetNASVolumeSnapshotsEntries": func() RestError {
			_, e := ep.GetNASVolumeSnapshotsEntries(ctx, "Pool-0", "d", 0, 0)
			return e
		},
		"GetNASSnapshotsEntries": func() RestError { _, e := ep.GetNASSnapshotsEntries(ctx, "Pool-0", 0, 0); return e },
		"CreateNASSnapshot": func() RestError {
			return ep.CreateNASSnapshot(ctx, "Pool-0", "d", &CreateNASSnapshotDescriptor{Name: "s"})
		},
		"DeleteNASSnapshot":    func() RestError { return ep.DeleteNASSnapshot(ctx, "Pool-0", "d", "s") },
		"GetNASVolumeSnapshot": func() RestError { _, e := ep.GetNASVolumeSnapshot(ctx, "Pool-0", "d", "s"); return e },
		"GetNASVolumeSnapshotClones": func() RestError {
			_, e := ep.GetNASVolumeSnapshotClones(ctx, "Pool-0", "d", "s")
			return e
		},
		"CreateNASClone": func() RestError {
			return ep.CreateNASClone(ctx, "Pool-0", "d", "s", CloneNASVolumeDescriptor{Name: "c"})
		},
		"DeleteNASClone": func() RestError { return ep.DeleteNASClone(ctx, "Pool-0", "d", "s", "c") },
		"CreateShare":    func() RestError { return ep.CreateShare(ctx, &CreateShareDescriptor{Name: "d", Path: "Pool-0/d/d"}) },
		"GetShare":       func() RestError { _, e := ep.GetShare(ctx, "d"); return e },
	}
}
