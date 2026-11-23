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
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

var noData = fakeReply{204, "application/json", nil}

// warnings returns the log lines at Warn that the endpoint wrote.
func (app *fakeAppliance) warnings() (lines []string) {
	for _, e := range app.log.entries {
		if e.Level == logrus.WarnLevel {
			lines = append(lines, e.Message)
		}
	}
	return lines
}

// A GET answered 204 carries no data. It is sent again after retrydelay,
// tries times in all, and the first reply with data is returned as if it
// had come first; a call that got its data after a 204 leaves one line at
// Warn.
func TestGetAnswered204IsSentAgain(t *testing.T) {
	app := newFakeAppliance(t)
	volume := fakeReply{200, "application/json", successBody(t, "GetVolume")}

	app.answers(noData, noData, volume)
	app.seen()
	vol, err := app.ep.GetVolume(app.ctx, "Pool-0", "v")
	if err != nil || vol == nil || vol.Name == "" {
		t.Fatalf("GetVolume answered 204, 204, 200: volume %v, error %v", vol, err)
	}
	if n := len(app.seen()); n != 3 {
		t.Errorf("%d requests sent, want 3", n)
	}
	warn := app.warnings()
	if len(warn) != 1 || !strings.Contains(warn[0], "GET api/v4/pools/Pool-0/volumes/v") ||
		!strings.Contains(warn[0], "status 200 after 2 replies of 204") {
		t.Errorf("the Warn lines of a GET made good after a 204: %q", warn)
	}

	// a reply with data to the first request leaves no line at Warn
	app.log.entries = nil
	app.answers(volume)
	if _, err := app.ep.GetVolume(app.ctx, "Pool-0", "v"); err != nil {
		t.Fatal(err.Error())
	}
	if warn := app.warnings(); len(warn) != 0 {
		t.Errorf("a GET answered at once left the Warn lines %q", warn)
	}
}

// A GET answered 204 to all of its requests is an error that names the
// status and the number of requests sent: never an empty object, an empty
// list or "no clones".
func TestGetAnswered204EveryTime(t *testing.T) {
	app := newFakeAppliance(t)
	app.answers(noData)
	app.seen()

	vol, err := app.ep.GetVolume(app.ctx, "Pool-0", "v")
	if vol != nil || err == nil {
		t.Fatalf("GetVolume answered 204 every time: volume %v, error %v", vol, err)
	}
	if err.GetCode() != RestErrorRPM || err.GetKindCode() != 0 ||
		!strings.Contains(err.Error(), "status 204") || !strings.Contains(err.Error(), "all 3 requests") {
		t.Errorf("error %q with the codes %d and %d", err.Error(), err.GetCode(), err.GetKindCode())
	}
	if n := len(app.seen()); n != 3 {
		t.Errorf("%d requests sent, want 3: the first and two more", n)
	}

	clones, err := app.ep.GetVolumeSnapshotClones(app.ctx, "Pool-0", "v", "s")
	if clones != nil || err == nil {
		t.Errorf("the clones of a snapshot answered 204: clones %v, error %v", clones, err)
	}
	page, err := app.ep.GetVolumesEntries(app.ctx, "Pool-0", 0, 0)
	if page != nil || err == nil {
		t.Errorf("a list answered 204: page %v, error %v", page, err)
	}
	pool, err := app.ep.GetPool(app.ctx, "Pool-0")
	if pool != nil || err == nil {
		t.Errorf("a pool answered 204: pool %v, error %v", pool, err)
	}
}

// A reply other than 204 to a repeated request is handled as if it had come
// first, an error of the catalogue included.
func TestGetAnswered204ThenAnError(t *testing.T) {
	app := newFakeAppliance(t)
	gone := `{"error": {"class": "zfslib.zfsapi.resources.ZfsResourceError", "errno": 1, "message": "Zfs resource: Pool-0/v not found in this collection."}}`
	app.answers(noData, fakeReply{500, "application/json", []byte(gone)})
	app.seen()
	_, err := app.ep.GetVolume(app.ctx, "Pool-0", "v")
	if err == nil || err.GetCode() != RestErrorResourceDNE {
		t.Fatalf("GetVolume answered 204, then not found: error %v", err)
	}
	if n := len(app.seen()); n != 2 {
		t.Errorf("%d requests sent, want 2", n)
	}
}

// Only a GET is sent again: a DELETE answered 204 has succeeded.
func TestDeleteAnswered204IsSentOnce(t *testing.T) {
	app := newFakeAppliance(t)
	app.answers(noData)
	app.seen()
	if err := app.ep.DeleteVolume(app.ctx, "Pool-0", "v", DeleteVolumeDescriptor{}); err != nil {
		t.Fatal(err.Error())
	}
	if n := len(app.seen()); n != 1 {
		t.Errorf("%d requests sent, want 1", n)
	}
}

// tries is 3 and retrydelay 1 s when the configuration does not set them;
// tries 1 means no second request; a retrydelay that is not a duration is
// refused by its name.
func TestTriesAndRetryDelay(t *testing.T) {
	l, _ := testLogger()
	cfg := jcom.RestEndpointCfg{Addrs: []string{"192.0.2.1"}, Port: 82, Prot: "https", IdleTimeOut: "30s"}
	var rp RestProxy
	if err := SetupRestProxy(&rp, &cfg, l); err != nil {
		t.Fatal(err)
	}
	if rp.tries != 3 || rp.retryDelay != time.Second {
		t.Errorf("without settings: tries %d, retrydelay %s; want 3 and 1s", rp.tries, rp.retryDelay)
	}

	cfg.Tries, cfg.RetryDelay = 5, "250ms"
	rp = RestProxy{}
	if err := SetupRestProxy(&rp, &cfg, l); err != nil {
		t.Fatal(err)
	}
	if rp.tries != 5 || rp.retryDelay != 250*time.Millisecond {
		t.Errorf("tries 5 and retrydelay 250ms read as %d and %s", rp.tries, rp.retryDelay)
	}

	cfg.RetryDelay = "soon"
	rp = RestProxy{}
	if err := SetupRestProxy(&rp, &cfg, l); err == nil || !strings.Contains(err.Error(), "retrydelay") {
		t.Errorf("retrydelay %q: error %v, want one naming retrydelay", cfg.RetryDelay, err)
	}

	// tries 1: the first request and no other
	app := newFakeAppliance(t)
	app.ep.rp.tries = 1
	app.answers(noData)
	app.seen()
	if _, err := app.ep.GetVolume(app.ctx, "Pool-0", "v"); err == nil || !strings.Contains(err.Error(), "all 1 requests") {
		t.Errorf("tries 1, answered 204: error %v", err)
	}
	if n := len(app.seen()); n != 1 {
		t.Errorf("tries 1: %d requests sent, want 1", n)
	}
}
