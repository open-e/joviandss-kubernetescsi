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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sessionTimeout = 30 * time.Second

// RestProxy - request client for any REST API
type RestProxy struct {
	addr          string
	port          int
	authToken     string
	httpRestProxy *http.Client
	l             *logrus.Entry
	prot          string
	user          string
	pass          string
	tries         int

	mu        sync.Mutex
	requestID int64
	timeout   int64
}

// RestProxyInterface - request client interface
type RestProxyInterface interface {
	Send(method, path string, data interface{}, ok int) (int, []byte, error)
}

func (rp *RestProxy) Send(method, path string, data interface{}, ok int) (int, []byte, error) {
	var res *http.Response
	var err error

	rp.mu.Lock()
	rp.requestID++
	rp.mu.Unlock()

	addr := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addr, rp.port, path)

	rp.l.Debug(fmt.Sprintf("Send %s request to %s", method, addr))

	// send request data as json
	var reader io.Reader
	if data == nil {
		reader = nil
	} else {
		jdata, err := json.Marshal(data)
		if err != nil {
			return 0, nil, err
		}
		reader = strings.NewReader(string(jdata))
	}

	//rp.l.Debugf("Url %+v", addr)

	req, err := http.NewRequest(method, addr, reader)
	req.SetBasicAuth(rp.user, rp.pass)
	if err != nil {
		rp.l.Warnf("Unable to create req: %s", err)
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kubernetes CSI Plugin")
	res, err = rp.httpRestProxy.Do(req)
	if err != nil {
		rp.l.Debugf("Request failed with error: %+v", err)
		return 0, nil, err
	}

	defer res.Body.Close()

	if err != nil {
		rp.l.Warnf("Request error: %+v", err)
		return 0, nil, err
	}

	// validate response body
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		rp.l.Warnf("Response failure: %+v", err)
		err = status.Error(codes.Internal, "Unable to process response")
		return res.StatusCode, nil, err
	}

	return res.StatusCode, bodyBytes, err
}

type RestProxyCfg struct {
	Addr        string
	Port        int
	Prot        string
	User        string
	Pass        string
	Pool        string
	IdleTimeOut string // See time Duration
	Tries       int
}

// TODO: implement sessions
func NewRestProxy(cfg RestProxyCfg, l *logrus.Entry) (ri RestProxyInterface, err error) {
	l = l.WithField("obj", "RestRestProxy")

	l.Debugf("Rest Proxy to %s", cfg.Addr)

	var timeoutDuration time.Duration

	timeoutDuration, err = time.ParseDuration(cfg.IdleTimeOut)

	if err != nil {
		l.Warnf("Uncorrect IdleTimeOut value: %s, Error %s", cfg.IdleTimeOut, err)
		return nil, err
	}

	tr := &http.Transport{
		IdleConnTimeout: sessionTimeout,
		TLSClientConfig: &tls.Config{
			// Connect without checking certificate
			InsecureSkipVerify: true,
		},
	}

	httpRestProxy := &http.Client{
		Transport: tr,
		Timeout:   timeoutDuration,
	}

	if cfg.Tries == 0 {
		cfg.Tries = 3
	}
	ri = &RestProxy{
		addr:          cfg.Addr,
		port:          cfg.Port,
		httpRestProxy: httpRestProxy,
		l:             l,
		requestID:     0,
		prot:          cfg.Prot,
		user:          cfg.User,
		pass:          cfg.Pass,
		tries:         cfg.Tries,
	}

	return ri, nil
}
