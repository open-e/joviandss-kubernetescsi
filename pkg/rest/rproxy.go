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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

const sessionTimeout = 30 * time.Second

// The settings a configuration may leave out.
const (
	defaultTries      = 3 // the first request and two more
	defaultRetryDelay = time.Second
)

// RestProxy - request client for any REST API
type RestProxy struct {
	addrs         []string
	active_addr   int
	port          int
	authToken     string
	httpRestProxy *http.Client
	setupErr      error // why the client was not set up; Send returns it
	l             *logrus.Entry
	prot          string
	user          string
	pass          string
	tries         int           // how many times a GET answered 204 is sent in all
	retryDelay    time.Duration // the pause before it is sent again

	mu        sync.Mutex
	requestID int64
	timeout   int64
}

// Send sends one request and reads its reply. It returns a RestError only
// when it has no reply to give back: no client, or no reply. A reply of any
// status comes back with its body.
func (rp *RestProxy) Send(ctx context.Context, method string, path string, data interface{}, ok int) (int, []byte, RestError) {
	if rp.httpRestProxy == nil {
		msg := "REST client not configured"
		if rp.setupErr != nil {
			msg += ": " + rp.setupErr.Error()
		}
		return 0, nil, GetError(RestErrorRequestMalfunction, msg)
	}

	// Neither the body nor a header of a request is ever logged: the body
	// can hold a CHAP password, and the Authorization header holds the REST
	// password.
	body := []byte("{}") // a GET and a DELETE carry it too
	if data != nil {
		var err error
		if body, err = json.Marshal(data); err != nil {
			return 0, nil, GetError(RestErrorRequestMalfunction, "request body: "+err.Error())
		}
	}

	url := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addrs[rp.active_addr], rp.port, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, GetError(RestErrorRequestMalfunction, "request: "+err.Error())
	}
	req.SetBasicAuth(rp.user, rp.pass) // only once the request exists
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kubernetes CSI Plugin")

	res, err := rp.httpRestProxy.Do(req)
	if err != nil {
		return 0, nil, transportError(err) // res is nil: no reply came
	}
	defer res.Body.Close()

	// The request log: one line for each reply, with the method, the path
	// with its query string and the status.
	jcom.LFC(ctx).WithFields(logrus.Fields{"func": "Send", "section": "rest"}).
		Debugf("%s %s: %d", method, path, res.StatusCode)

	reply, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, GetError(RestErrorRequestMalfunction, "reading the reply: "+err.Error())
	}
	return res.StatusCode, reply, nil
}

// transportError names the kind of failure of a request that got no reply.
// It never reads the reply that did not come.
func transportError(err error) RestError {
	var certErr *tls.CertificateVerificationError
	var dnsErr *net.DNSError
	var netErr net.Error
	var opErr *net.OpError
	switch {
	case errors.As(err, &certErr):
		return GetError(RestErrorUnableToConnect, fmt.Sprintf(
			"certificate of the appliance not verified: %s; set cafile to the certificate or its CA, or set insecure",
			certErr.Err))
	case errors.As(err, &dnsErr):
		return GetError(RestErrorUnableToConnect, "address does not resolve: "+dnsErr.Error())
	case errors.As(err, &netErr) && netErr.Timeout(): // the client's own timeout too
		return GetError(RestErrorRequestTimeout, "request timed out: "+err.Error())
	case errors.As(err, &opErr) && opErr.Op == "dial":
		return GetError(RestErrorUnableToConnect, "connection refused or failed: "+opErr.Error())
	default:
		return GetError(RestErrorRequestMalfunction, "request failed: "+err.Error())
	}
}

// getWithData sends a GET of the plugin. A reply of 204 carries no data,
// so the same GET is sent again after retrydelay, tries times in all. A
// 204 is never turned into an empty object or an empty list: a caller
// would read those as a volume of size 0, a full pool or a snapshot
// without clones. Only a GET is sent again: a DELETE answered 204 has
// succeeded.
func (rp *RestProxy) getWithData(ctx context.Context, path string) (int, []byte, RestError) {
	for sent := 1; ; sent++ {
		status, body, err := rp.Send(ctx, "GET", path, nil, CodeOK)
		if err != nil || status != CodeNoContent {
			if err == nil && status == CodeOK && sent > 1 {
				// no recording shows a 204 on a GET: one made good leaves a trace
				jcom.LFC(ctx).Warnf("GET %s: status %d after %d replies of 204", path, status, sent-1)
			}
			return status, body, err
		}
		if sent >= rp.tries {
			return status, nil, GetError(RestErrorRPM,
				fmt.Sprintf("GET %s: status 204, no data, to all %d requests", path, sent))
		}
		time.Sleep(rp.retryDelay)
	}
}

// SetupRestProxy checks the endpoint configuration and builds the client.
// Its error names the setting it refuses.
func SetupRestProxy(rp *RestProxy, cfg *jcom.RestEndpointCfg, l *logrus.Entry) error {
	rp.l = l.WithFields(logrus.Fields{"section": "restproxy", "addrs": cfg.Addrs, "port": cfg.Port})

	if len(cfg.Addrs) == 0 {
		return fmt.Errorf("addrs: no address of the appliance is set")
	}
	if cfg.Prot != "" && cfg.Prot != "https" {
		return fmt.Errorf("prot: %q is refused, the plugin uses https only", cfg.Prot)
	}
	timeout, err := time.ParseDuration(cfg.IdleTimeOut)
	if err != nil {
		return fmt.Errorf("idletimeout: %w", err)
	}
	retryDelay := defaultRetryDelay
	if cfg.RetryDelay != "" {
		if retryDelay, err = time.ParseDuration(cfg.RetryDelay); err != nil {
			return fmt.Errorf("retrydelay: %w", err)
		}
	}

	// The certificate of the appliance is checked unless insecure is set,
	// and nothing below TLS 1.2 is spoken.
	tc := &tls.Config{MinVersion: tls.VersionTLS12}
	switch {
	case cfg.Insecure:
		tc.InsecureSkipVerify = true // no check at all; cafile is not read
	case cfg.CAFile != "":
		pem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return fmt.Errorf("cafile: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil || roots == nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(pem) {
			return fmt.Errorf("cafile: %s holds no PEM certificate", cfg.CAFile)
		}
		// the appliance's own certificate, or the CA that signed it; the
		// certificate must still be within its dates and list the address
		tc.RootCAs = roots
	}
	rp.httpRestProxy = &http.Client{
		Transport: &http.Transport{IdleConnTimeout: sessionTimeout, TLSClientConfig: tc},
		Timeout:   timeout,
		// The API has no reason to redirect: a redirect means that something
		// else answered, and its reply is returned as it came.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	rp.prot = "https"
	rp.addrs = append(rp.addrs[:0], cfg.Addrs...)
	rp.active_addr = 0
	rp.port = cfg.Port
	rp.requestID = 0
	rp.user = cfg.User
	rp.pass = cfg.Pass
	rp.tries = cfg.Tries
	if rp.tries <= 0 {
		rp.tries = defaultTries
	}
	rp.retryDelay = retryDelay

	return nil
}
