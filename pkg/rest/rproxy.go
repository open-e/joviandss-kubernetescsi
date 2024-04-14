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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	httpUrl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "joviandss-kubernetescsi/pkg/common"
)

const sessionTimeout = 30 * time.Second

// RestProxy - request client for any REST API
type RestProxy struct {
	addrs         []string
	active_addr   int
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

// restProxy - request client for any REST API
type restProxy struct {
	addrs         []string
	active_addr   int
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
	//Send(method, path string, data interface{}, ok int) (int, []byte, error)
	Send(ctx context.Context, method string, path string, data interface{}, ok int) (int, []byte, RestError)
}

func (rp *RestProxy) Send(ctx context.Context, method string, path string, data interface{}, ok int) (int, []byte, RestError) {
	var res *http.Response
	// var err restError
	l := jcom.LFC(ctx)
	l.Debugf("Path %s", path)
	//l.Debugf("proto %+v, addr %+v, port %+v, url %+v", )

	url := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addrs[rp.active_addr], rp.port, path)

	l = l.WithFields(logrus.Fields{
		"func":    "Send",
		"section": "rest",
		"method":  method,
		"url":     url,
	})

	l.Debugf("Available addrs %+v", rp.addrs)

	//rp.mu.Lock()
	//rp.requestID++
	//rp.mu.Unlock()

	// send request data as json
	var reader io.Reader
	if data == nil {
		l.Debug("sending with no data")
		reader = nil
	} else {
		l.Debugf("sending data %+v", data)
		jdata, err := json.Marshal(data)
		if err != nil {
			return 0, nil, &restError{RestErrorRequestMalfunction, err.Error()}
		}
		l.Debugf("sending marshaled data %s", jdata)
		reader = strings.NewReader(string(jdata))
	}

	req, err := http.NewRequest(method, url, reader)
	req.SetBasicAuth(rp.user, rp.pass)
	if err != nil {
		//rp.l.Warnf("Unable to create req: %s", err)
		return 0, nil, &restError{RestErrorRequestMalfunction, err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kubernetes CSI Plugin")
	res, err = rp.httpRestProxy.Do(req)

	if err != nil {
		if urlErr, ok := err.(*httpUrl.Error); ok {
			if opErr, ok := urlErr.Err.(*net.OpError); ok {
				if dnsErr, ok := opErr.Err.(*net.DNSError); ok {
					l.Errorln("DNS error:", dnsErr)
					return res.StatusCode, nil, &restError{RestErrorUnableToConnect, dnsErr.Error()}
				} else if opErr.Op == "dial" {
					l.Errorln("Connection error:", opErr)
					return res.StatusCode, nil, &restError{RestErrorUnableToConnect, opErr.Error()}
				} else if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						l.Errorln("Network error (timeout):", netErr.Error())
						return res.StatusCode, nil, &restError{RestErrorRequestTimeout, opErr.Error()}
					} else {
						l.Errorln("Network error:", netErr)
						return res.StatusCode, nil, &restError{RestErrorRequestMalfunction, opErr.Error()}
					}
				}
			} else {
				fmt.Println("Network error:", opErr)
				return res.StatusCode, nil, &restError{RestErrorRequestMalfunction, urlErr.Error()}
			}
		} else {
			fmt.Println("Unknown error:", err.Error())
			return res.StatusCode, nil, &restError{RestErrorRequestMalfunction, err.Error()}
		}
	}

	defer res.Body.Close()

	// validate response body
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		l.Errorf("reading response failed: %s", err.Error())
		err = status.Error(codes.Internal, "Unable to process response")
		return res.StatusCode, nil, &restError{RestErrorRequestMalfunction, err.Error()}
	}
	l.Debugf("Request completed with code %d, obtained %d bytes", res.StatusCode, len(bodyBytes))
	return res.StatusCode, bodyBytes, nil
}


type RestProxyCfg struct {
	Addrs       string
	ActiveAddr  int
	Port        int
	Prot        string
	User        string
	Pass        string
	Pool        string
	IdleTimeOut string // See time Duration
	Tries       int
}

// TODO: implement sessions
// func NewRestProxy(cfg *RestEndpointCfg, l *logrus.Entry) (ri RestProxyInterface, err error) {
//
//
// 	var timeoutDuration time.Duration
//
// 	le := l.WithFields(logrus.Fields{"section": "restproxy", "addrs": cfg.Addrs, "port": cfg.Port,})
//
// 	l.Debugf("Rest Proxy to %v", cfg.Addrs)
//
// 	timeoutDuration, err = time.ParseDuration(cfg.IdleTimeOut)
//
// 	if err != nil {
// 		l.Warnf("Uncorrect IdleTimeOut value: %s, Error %s", cfg.IdleTimeOut, err)
// 		return nil, err
// 	}
//
// 	tr := &http.Transport{
// 		IdleConnTimeout: sessionTimeout,
// 		TLSClientConfig: &tls.Config{
// 			// Connect without checking certificate
// 			InsecureSkipVerify: true,
// 		},
// 	}
//
// 	httpRestProxy := &http.Client{
// 		Transport: tr,
// 		Timeout:   timeoutDuration,
// 	}
//
// 	if cfg.Tries == 0 {
// 		cfg.Tries = 3
// 	}
// 	ri = &RestProxy{
// 		addrs:          cfg.Addrs,
// 		active_addr:	0,
// 		port:		cfg.Port,
// 		httpRestProxy:	httpRestProxy,
// 		l:             	le,
// 		requestID:     	0,
// 		prot:          	cfg.Prot,
// 		user:          	cfg.User,
// 		pass:          	cfg.Pass,
// 		tries:         	cfg.Tries,
// 	}
//
// 	return ri, nil
// }

func SetupRestProxy(rp *RestProxy, cfg *jcom.RestEndpointCfg, l *logrus.Entry) (err error) {

	// rp.l = l.WithField("section", "restproxy")
	rp.l = l.WithFields(logrus.Fields{"section": "restproxy", "addrs": cfg.Addrs, "port": cfg.Port})

	rp.l.Debug("Setting up rest proxy")
	var timeoutDuration time.Duration

	timeoutDuration, err = time.ParseDuration(cfg.IdleTimeOut)

	if err != nil {
		logrus.Warnf("Uncorrect IdleTimeOut value: %s, Error %s", cfg.IdleTimeOut, err)
		return err
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
	//*rp = RestProxy{
	//	active_addr:	0,
	//	port:		cfg.Port,
	//	httpRestProxy:	httpRestProxy,
	//	requestID:     	0,
	//	prot:          	cfg.Prot,
	//	user:          	cfg.User,
	//	pass:          	cfg.Pass,
	//	tries:         	cfg.Tries,
	//}
	rp.addrs = append(rp.addrs, cfg.Addrs...)
	rp.active_addr = 0
	rp.port = cfg.Port
	rp.httpRestProxy = httpRestProxy
	rp.requestID = 0
	rp.prot = cfg.Prot
	rp.user = cfg.User
	rp.pass = cfg.Pass
	rp.tries = cfg.Tries

	return nil
}
