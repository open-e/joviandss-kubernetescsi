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
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	page, err := app.ep.GetVolumesEntries(app.ctx, "Pool-0", 0, 100)
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

// endpointFor sets up an endpoint for the address of a test server, with
// the client timeout given.
func endpointFor(t *testing.T, serverURL string, timeout string) (*RestEndpoint, context.Context) {
	t.Helper()
	u, err := url.Parse(serverURL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	l, _ := testLogger()
	cfg := jcom.RestEndpointCfg{Addrs: []string{u.Hostname()}, Port: port, User: "admin", Pass: testRestPassword,
		IdleTimeOut: timeout, RetryDelay: "1ms", Insecure: true}
	var ep RestEndpoint
	if err := SetupEndpoint(&ep, &cfg, l); err != nil {
		t.Fatal(err)
	}
	return &ep, withTestLogger(context.Background(), l)
}

// A request that gets no reply returns a RestError that names the kind of
// its failure, with 0 as its second code: nothing reads the reply that did
// not come.
func TestRequestWithoutAReply(t *testing.T) {
	// nothing listens on the port: the connection is refused
	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	ep, ctx := endpointFor(t, srv.URL, "30s")
	srv.Close()
	vol, err := ep.GetVolume(ctx, "Pool-0", "v")
	if vol != nil || err == nil || err.GetCode() != RestErrorUnableToConnect || err.GetKindCode() != 0 {
		t.Fatalf("a refused connection: volume %v, error %v", vol, err)
	}
	if !strings.Contains(err.Error(), "connection refused or failed") {
		t.Errorf("a refused connection: %q", err.Error())
	}
	if derr := ep.DeleteVolume(ctx, "Pool-0", "v", DeleteVolumeDescriptor{}); derr == nil || derr.GetCode() != RestErrorUnableToConnect {
		t.Errorf("a refused connection on a delete: %v", derr)
	}

	// the appliance does not answer in time: the client's own timeout
	release := make(chan struct{})
	slow := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	defer slow.Close()
	defer close(release)
	ep, ctx = endpointFor(t, slow.URL, "100ms")
	_, err = ep.GetPool(ctx, "Pool-0")
	if err == nil || err.GetCode() != RestErrorRequestTimeout || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("a request without an answer in time: %v", err)
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// transportError tells the kinds of a failure without a reply apart.
func TestTransportErrorKinds(t *testing.T) {
	dial := func(err error) error {
		return &url.Error{Op: "Get", URL: "https://jdss:82/api/v4/pools/Pool-0", Err: &net.OpError{Op: "dial", Net: "tcp", Err: err}}
	}
	for _, c := range []struct {
		what string
		err  error
		code int
		text string
	}{
		{"an address that does not resolve", dial(&net.DNSError{Err: "no such host", Name: "jdss", IsNotFound: true}), RestErrorUnableToConnect, "does not resolve"},
		{"a refused connection", dial(syscall.ECONNREFUSED), RestErrorUnableToConnect, "connection refused or failed"},
		{"a dial that times out", dial(timeoutError{}), RestErrorRequestTimeout, "timed out"},
		{"the client's own timeout", &url.Error{Op: "Get", URL: "https://jdss:82/", Err: timeoutError{}}, RestErrorRequestTimeout, "timed out"},
		{"a connection reset while the reply is read", &url.Error{Op: "Get", URL: "https://jdss:82/", Err: &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}}, RestErrorRequestMalfunction, "request failed"},
		{"anything else", errors.New("unexpected EOF"), RestErrorRequestMalfunction, "request failed"},
	} {
		err := transportError(c.err)
		if err.GetCode() != c.code || err.GetKindCode() != 0 || !strings.Contains(err.Error(), c.text) {
			t.Errorf("%s: %q with the codes %d and %d, want %q with the code %d", c.what, err.Error(), err.GetCode(), err.GetKindCode(), c.text, c.code)
		}
	}
}

// A request that cannot be built is a RestError: the authentication is
// added only once the request exists.
func TestRequestThatCannotBeBuilt(t *testing.T) {
	app := newFakeAppliance(t)
	app.seen()
	for _, c := range []struct{ method, path string }{
		{"GET", "api/v4/pools/Pool\x7f0"},
		{"BAD METHOD", "api/v4/pools/Pool-0"},
	} {
		status, body, err := app.ep.rp.Send(app.ctx, c.method, c.path, nil, CodeOK)
		if err == nil || err.GetCode() != RestErrorRequestMalfunction || status != 0 || body != nil {
			t.Errorf("%s %q: status %d, error %v", c.method, c.path, status, err)
		}
	}
	// a request body that does not encode
	if _, _, err := app.ep.rp.Send(app.ctx, "POST", "api/v4/shares", map[string]interface{}{"f": func() {}}, CodeOK); err == nil || err.GetCode() != RestErrorRequestMalfunction {
		t.Errorf("a body that does not encode: %v", err)
	}
	if n := len(app.seen()); n != 0 {
		t.Errorf("%d requests reached the appliance", n)
	}
}

// A configuration the client cannot use is refused by the name of its
// setting; the error is returned, logged once, and returned again by every
// REST call, which never panics.
func TestSetupErrorIsReturnedByEveryCall(t *testing.T) {
	good := jcom.RestEndpointCfg{Addrs: []string{"192.0.2.1"}, Port: 82, Prot: "https", IdleTimeOut: "30s"}
	for _, c := range []struct {
		setting string
		change  func(cfg *jcom.RestEndpointCfg)
	}{
		{"addrs", func(cfg *jcom.RestEndpointCfg) { cfg.Addrs = nil }},
		{"prot", func(cfg *jcom.RestEndpointCfg) { cfg.Prot = "http" }},
		{"idletimeout", func(cfg *jcom.RestEndpointCfg) { cfg.IdleTimeOut = "soon" }},
		{"idletimeout", func(cfg *jcom.RestEndpointCfg) { cfg.IdleTimeOut = "" }},
		{"retrydelay", func(cfg *jcom.RestEndpointCfg) { cfg.RetryDelay = "later" }},
	} {
		cfg := good
		c.change(&cfg)
		l, rec := testLogger()
		var ep RestEndpoint
		err := SetupEndpoint(&ep, &cfg, l)
		if err == nil || !strings.HasPrefix(err.Error(), c.setting+":") {
			t.Errorf("a bad %s: setup error %v, want one that names the setting", c.setting, err)
			continue
		}
		errors := 0
		for _, e := range rec.entries {
			if e.Level == logrus.ErrorLevel && strings.Contains(e.Message, c.setting) {
				errors++
			}
		}
		if errors != 1 {
			t.Errorf("a bad %s: %d Error lines that name it, want 1", c.setting, errors)
		}
		ctx := withTestLogger(context.Background(), l)
		_, rerr := ep.GetPool(ctx, "Pool-0")
		if rerr == nil || rerr.GetCode() != RestErrorRequestMalfunction ||
			!strings.Contains(rerr.Error(), "REST client not configured") || !strings.Contains(rerr.Error(), c.setting) {
			t.Errorf("a call after a bad %s: %v", c.setting, rerr)
		}
		if rerr := ep.CreateVolume(ctx, "Pool-0", CreateVolumeDescriptor{Name: "v"}); rerr == nil || !strings.Contains(rerr.Error(), c.setting) {
			t.Errorf("a create after a bad %s: %v", c.setting, rerr)
		}
	}

	// an endpoint that was never set up at all
	var never RestEndpoint
	l, _ := testLogger()
	if _, rerr := never.GetPool(withTestLogger(context.Background(), l), "Pool-0"); rerr == nil || !strings.Contains(rerr.Error(), "REST client not configured") {
		t.Errorf("a call on an endpoint that was never set up: %v", rerr)
	}
}

// A configuration without prot works as HTTPS.
func TestProtIsHTTPSWhenItIsNotSet(t *testing.T) {
	app := newFakeAppliance(t) // an HTTPS server
	ep, ctx := endpointFor(t, app.srv.URL, "30s")
	if ep.rp.prot != "https" {
		t.Errorf("without prot the scheme is %q", ep.rp.prot)
	}
	app.answer(200, "application/json", successBody(t, "GetPool"))
	if _, err := ep.GetPool(ctx, "Pool-0"); err != nil {
		t.Errorf("a call of an endpoint without prot: %s", err.Error())
	}
}

// An endpoint formats with %v and %s without a panic, with addresses and
// without.
func TestEndpointString(t *testing.T) {
	var empty RestEndpoint
	if got := fmt.Sprintf("%v %s", &empty, &empty); !strings.Contains(got, "port: 0") {
		t.Errorf("an endpoint without addresses formats as %q", got)
	}
	ep := RestEndpoint{rec: jcom.RestEndpointCfg{Addrs: []string{"192.0.2.1", "192.0.2.2"}, Port: 82}}
	if got := ep.String(); got != "addrs: 192.0.2.1,192.0.2.2 port: 82" {
		t.Errorf("an endpoint formats as %q", got)
	}
}

// tlsEndpoint sets up an endpoint for a TLS test server with the settings a
// test changes.
func tlsEndpoint(t *testing.T, srv *httptest.Server, change func(cfg *jcom.RestEndpointCfg)) (*RestEndpoint, context.Context, error) {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(u.Port())
	l, _ := testLogger()
	cfg := jcom.RestEndpointCfg{Addrs: []string{u.Hostname()}, Port: port, User: "admin", Pass: testRestPassword, IdleTimeOut: "30s"}
	if change != nil {
		change(&cfg)
	}
	var ep RestEndpoint
	serr := SetupEndpoint(&ep, &cfg, l)
	return &ep, withTestLogger(context.Background(), l), serr
}

// poolServer answers every request with a recorded pool reply.
func poolServer(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	body := successBody(t, "GetPool")
	requests := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &requests
}

// The certificate of the appliance is checked by default. With neither
// setting, an appliance whose certificate does not verify is refused with
// an error that names the certificate problem and the two settings; with
// cafile naming the certificate, or with insecure, the call succeeds.
func TestCertificateIsCheckedByDefault(t *testing.T) {
	srv, requests := poolServer(t) // its certificate is signed by no known CA

	ep, ctx, err := tlsEndpoint(t, srv, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, rerr := ep.GetPool(ctx, "Pool-0")
	if rerr == nil || rerr.GetCode() != RestErrorUnableToConnect {
		t.Fatalf("a certificate that does not verify: error %v", rerr)
	}
	for _, part := range []string{"certificate", "not verified", "cafile", "insecure"} {
		if !strings.Contains(rerr.Error(), part) {
			t.Errorf("the error %q does not say %q", rerr.Error(), part)
		}
	}
	if *requests != 0 {
		t.Errorf("%d requests reached an appliance whose certificate did not verify", *requests)
	}

	// cafile: the certificate of the appliance itself
	cafile := filepath.Join(t.TempDir(), "appliance.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(cafile, pemBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	ep, ctx, err = tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.CAFile = cafile })
	if err != nil {
		t.Fatal(err)
	}
	if _, rerr := ep.GetPool(ctx, "Pool-0"); rerr != nil {
		t.Errorf("with cafile naming the certificate: %s", rerr.Error())
	}

	// insecure: no check
	ep, ctx, err = tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.Insecure = true })
	if err != nil {
		t.Fatal(err)
	}
	if _, rerr := ep.GetPool(ctx, "Pool-0"); rerr != nil {
		t.Errorf("with insecure: %s", rerr.Error())
	}
	// insecure wins: cafile is not read
	if _, _, err = tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.Insecure, cfg.CAFile = true, "/no/such/file" }); err != nil {
		t.Errorf("insecure with a cafile that does not exist: %v", err)
	}

	// cafile cannot make up for an address the certificate does not list
	other := filepath.Join(t.TempDir(), "other.pem")
	otherSrv, _ := poolServer(t)
	otherPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: otherSrv.Certificate().Raw})
	os.WriteFile(other, otherPEM, 0o600)
	ep, ctx, _ = tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) {
		cfg.CAFile = other
		cfg.Addrs = []string{"localhost"} // the test certificate lists 127.0.0.1 and example.com
	})
	if _, rerr := ep.GetPool(ctx, "Pool-0"); rerr == nil || !strings.Contains(rerr.Error(), "not verified") {
		t.Errorf("an address the certificate does not list: error %v", rerr)
	}
}

// A cafile that cannot be read, or holds no certificate, is refused by its
// name at setup.
func TestCAFileThatCannotBeUsed(t *testing.T) {
	srv, _ := poolServer(t)
	empty := filepath.Join(t.TempDir(), "empty.pem")
	os.WriteFile(empty, []byte("not a certificate\n"), 0o600)
	for what, file := range map[string]string{"a file that does not exist": "/no/such/file.pem", "a file without a certificate": empty} {
		ep, ctx, err := tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.CAFile = file })
		if err == nil || !strings.HasPrefix(err.Error(), "cafile:") {
			t.Errorf("%s: setup error %v, want one that names cafile", what, err)
		}
		if _, rerr := ep.GetPool(ctx, "Pool-0"); rerr == nil || !strings.Contains(rerr.Error(), "cafile") {
			t.Errorf("%s: a call returns %v", what, rerr)
		}
	}
}

// A server that offers less than TLS 1.2 is refused, with insecure too.
func TestTLSBelow12IsRefused(t *testing.T) {
	requests := 0
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	srv.TLS = &tls.Config{MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS11}
	srv.StartTLS()
	defer srv.Close()
	ep, ctx, err := tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.Insecure = true })
	if err != nil {
		t.Fatal(err)
	}
	if _, rerr := ep.GetPool(ctx, "Pool-0"); rerr == nil || requests != 0 {
		t.Errorf("a server of TLS 1.1 at most: error %v, %d requests served", rerr, requests)
	}
	if ep.rp.httpRestProxy.Transport.(*http.Transport).TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Error("the client's minimum is not TLS 1.2")
	}
}

// A redirect is not followed: the API has no reason to redirect, so the
// reply is returned as it came, and is an error of the call.
func TestRedirectIsNotFollowed(t *testing.T) {
	target, followed := poolServer(t)
	redirects := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirects++
		http.Redirect(w, r, target.URL+"/api/v4/pools/Pool-0", http.StatusFound)
	}))
	defer srv.Close()
	ep, ctx, err := tlsEndpoint(t, srv, func(cfg *jcom.RestEndpointCfg) { cfg.Insecure = true })
	if err != nil {
		t.Fatal(err)
	}
	pool, rerr := ep.GetPool(ctx, "Pool-0")
	if pool != nil || rerr == nil || !strings.Contains(rerr.Error(), "status 302") {
		t.Errorf("a redirect: pool %v, error %v", pool, rerr)
	}
	if redirects != 1 || *followed != 0 {
		t.Errorf("%d requests to the appliance and %d to the place it named, want 1 and 0", redirects, *followed)
	}
}

// A reply larger than maxreplysize fails with an error that names the
// limit, and one of exactly the limit is read whole.
func TestReplySizeLimit(t *testing.T) {
	app := newFakeAppliance(t)
	app.ep.rp.maxReplySize = 1000
	body := func(n int) []byte {
		// a volume reply, padded to n bytes with a field the plugin does not read
		head := `{"data": {"name": "v", "volsize": "1", "guid": "g", "default_scsi_id": "s", "comment": "`
		tail := `"}}`
		return []byte(head + strings.Repeat("x", n-len(head)-len(tail)) + tail)
	}

	app.answer(200, "application/json", body(1000))
	if vol, err := app.ep.GetVolume(app.ctx, "Pool-0", "v"); err != nil || vol.Name != "v" {
		t.Errorf("a reply of exactly the limit: %v", err)
	}
	app.answer(200, "application/json", body(1001))
	vol, err := app.ep.GetVolume(app.ctx, "Pool-0", "v")
	if vol != nil || err == nil || err.GetCode() != RestErrorRPM || !strings.Contains(err.Error(), "maxreplysize, 1000 bytes") {
		t.Errorf("a reply one byte over the limit: volume %v, error %v", vol, err)
	}
	// an error reply over the limit is not read into memory either
	app.answer(500, "application/json", body(5000))
	if err := app.ep.DeleteVolume(app.ctx, "Pool-0", "v", DeleteVolumeDescriptor{}); err == nil || !strings.Contains(err.Error(), "maxreplysize") {
		t.Errorf("an error reply over the limit: %v", err)
	}
}

// A reply without an end is read no further than the limit: the call
// returns, and the appliance is not listened to any longer.
func TestReplyWithoutAnEndIsNotReadWhole(t *testing.T) {
	written := make(chan int, 1)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		chunk := []byte(strings.Repeat("x", 64*1024))
		n := 0
		for n < 1<<30 { // a gigabyte, unless the client hangs up
			if _, err := w.Write(chunk); err != nil {
				break
			}
			n += len(chunk)
		}
		written <- n
	}))
	defer srv.Close()
	ep, ctx := endpointFor(t, srv.URL, "30s")
	done := make(chan RestError, 1)
	go func() {
		_, err := ep.GetPool(ctx, "Pool-0")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil || err.GetCode() != RestErrorRPM || !strings.Contains(err.Error(), "maxreplysize, 1048576 bytes") {
			t.Errorf("a reply without an end: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the call is still reading the reply")
	}
	select {
	case n := <-written:
		if n > 64<<20 {
			t.Errorf("the appliance sent %d bytes before the plugin hung up", n)
		}
	case <-time.After(20 * time.Second):
		t.Error("the appliance is still sending: the connection was not closed")
	}
}

// maxreplysize is 1 MiB when the configuration does not set it, and a
// negative value is refused by its name.
func TestMaxReplySizeSetting(t *testing.T) {
	l, _ := testLogger()
	cfg := jcom.RestEndpointCfg{Addrs: []string{"192.0.2.1"}, Port: 82, IdleTimeOut: "30s"}
	var rp RestProxy
	if err := SetupRestProxy(&rp, &cfg, l); err != nil || rp.maxReplySize != 1<<20 {
		t.Errorf("without the setting: %d bytes, error %v; want 1048576", rp.maxReplySize, err)
	}
	cfg.MaxReplySize = 4096
	rp = RestProxy{}
	if err := SetupRestProxy(&rp, &cfg, l); err != nil || rp.maxReplySize != 4096 {
		t.Errorf("maxreplysize 4096 is read as %d, error %v", rp.maxReplySize, err)
	}
	cfg.MaxReplySize = -1
	rp = RestProxy{}
	if err := SetupRestProxy(&rp, &cfg, l); err == nil || !strings.HasPrefix(err.Error(), "maxreplysize:") {
		t.Errorf("maxreplysize -1: error %v, want one that names the setting", err)
	}
}
