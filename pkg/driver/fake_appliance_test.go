package driver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// fakeRequest is one request the fake appliance got.
type fakeRequest struct {
	Method string
	Path   string // without the query
	Query  url.Values
	Body   map[string]interface{}
	raw    string // the query as it was sent
}

// line writes a request as the tests compare it: the method, the path and
// the query as it was sent.
func (r fakeRequest) line() string {
	if len(r.Query) == 0 {
		return r.Method + " " + r.Path
	}
	return r.Method + " " + r.Path + "?" + r.raw
}

// fakeAppliance is an HTTPS server that answers the requests of a driver
// from the routes a test gives it. A request without a route is answered
// 404 without a JSON body and fails the test.
type fakeAppliance struct {
	t      *testing.T
	srv    *httptest.Server
	ctx    context.Context
	cfg    jcom.RestEndpointCfg
	logger *logrus.Entry
	warned []string

	mu       sync.Mutex
	routes   map[string]func(r fakeRequest) (int, interface{})
	requests []fakeRequest
}

type recordWarnings struct{ app *fakeAppliance }

func (recordWarnings) Levels() []logrus.Level { return []logrus.Level{logrus.WarnLevel} }

func (h recordWarnings) Fire(e *logrus.Entry) error {
	h.app.warned = append(h.app.warned, e.Message)
	return nil
}

func newFakeAppliance(t *testing.T) *fakeAppliance {
	t.Helper()
	app := &fakeAppliance{t: t, routes: map[string]func(fakeRequest) (int, interface{}){}}
	app.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		req := fakeRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query()}
		req.raw = r.URL.RawQuery
		json.Unmarshal(raw, &req.Body)
		app.mu.Lock()
		app.requests = append(app.requests, req)
		route := app.routes[r.Method+" "+r.URL.Path]
		app.mu.Unlock()
		if route == nil {
			t.Errorf("the driver sent a request this test has no answer for: %s", req.line())
			w.WriteHeader(404)
			w.Write([]byte("404 not found"))
			return
		}
		status, reply := route(req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if reply != nil {
			json.NewEncoder(w).Encode(reply)
		}
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
	lg := logrus.New()
	lg.SetOutput(io.Discard)
	lg.SetLevel(logrus.DebugLevel)
	lg.AddHook(recordWarnings{app})
	app.logger = logrus.NewEntry(lg)
	app.cfg = jcom.RestEndpointCfg{
		Addrs: []string{u.Hostname()}, Port: port, Prot: "https",
		User: "admin", Pass: "admin", IdleTimeOut: "30s", RetryDelay: "1ms",
	}
	app.ctx = jcom.WithLogger(context.Background(), app.logger)
	return app
}

// iscsi and nfs give a driver of each protocol that talks to the fake
// appliance.
func (app *fakeAppliance) iscsi() *CSIISCSiDriver {
	var d CSIISCSiDriver
	if err := jrest.SetupEndpoint(&d.re, &app.cfg, app.logger); err != nil {
		app.t.Fatal(err)
	}
	return &d
}

func (app *fakeAppliance) nfs() *CSINFSDriver {
	var d CSINFSDriver
	if err := jrest.SetupEndpoint(&d.re, &app.cfg, app.logger); err != nil {
		app.t.Fatal(err)
	}
	return &d
}

// on sets the answer to a method and a path.
func (app *fakeAppliance) on(method, path string, route func(r fakeRequest) (int, interface{})) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.routes[method+" "+path] = route
}

// seen returns the requests got so far, as lines, and forgets them.
func (app *fakeAppliance) seen() (lines []string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	for _, r := range app.requests {
		lines = append(lines, r.line())
	}
	app.requests = nil
	return lines
}

// sent returns the requests got so far and forgets them.
func (app *fakeAppliance) sent() []fakeRequest {
	app.mu.Lock()
	defer app.mu.Unlock()
	out := app.requests
	app.requests = nil
	return out
}

type object = map[string]interface{}

// data is the reply of a success with data.
func data(v interface{}) object { return object{"data": v, "error": nil} }

// appError is the reply of a failure: class "" and errno -1 stand for a
// field the appliance does not send.
func appError(class string, errno int, message string) object {
	e := object{"message": message, "url": "http://127.0.0.1:81/"}
	if class != "" {
		e["class"] = class
	}
	if errno >= 0 {
		e["errno"] = errno
	}
	return object{"data": nil, "error": e}
}

// pages answers a list request from entries, page by page, as the
// appliance does.
func pages(entries func() []object) func(r fakeRequest) (int, interface{}) {
	return func(r fakeRequest) (int, interface{}) {
		all := entries()
		page, _ := strconv.Atoi(r.Query.Get("page"))
		per, err := strconv.Atoi(r.Query.Get("per_page"))
		if err != nil || per <= 0 {
			per = 10 // an appliance's own page size: a request must not rely on it
		}
		from := page * per
		if from > len(all) {
			from = len(all)
		}
		to := from + per
		if to > len(all) {
			to = len(all)
		}
		return 200, data(object{"results": len(all), "entries": all[from:to]})
	}
}

// volumeObject is an entry of the volume list, and the reply of a volume.
func volumeObject(name string) object {
	return object{"name": name, "volsize": "1073741824", "origin": nil, "is_clone": false,
		"guid": "guid-" + name, "default_scsi_id": "scsi-" + name}
}

// nasVolumeObject is an entry of the NAS volume list.
func nasVolumeObject(name string) object {
	return object{"name": name, "quota": "1073741824", "origin": nil}
}

// snapshotObject is a short entry of a snapshot list. kind is the kind of
// its parent: 4 for a zvol, 1 for a NAS volume.
func snapshotObject(parent, name string, kind int) object {
	return object{"name": name,
		"parent":     object{"name": parent, "type": kind},
		"properties": object{"creation": 1762000000, "guid": "guid-" + parent + "@" + name}}
}
