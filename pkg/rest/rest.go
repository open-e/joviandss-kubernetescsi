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
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

var vendorVersion = "dev"

// apiPrefix is the one prefix of every path the plugin requests. The plugin
// uses JovianDSS REST API v4 only: there is no setting for the version and
// no fallback to an older one.
const apiPrefix = "api/v4"

type RestEndpoint struct {
	rec jcom.RestEndpointCfg
	rp  RestProxy
	l   *logrus.Entry
}

// String names the endpoint in a log line. Go calls it whenever the
// endpoint is formatted with %v or %s, so it reads nothing that may be
// missing.
func (re *RestEndpoint) String() string {
	return fmt.Sprintf("addrs: %s port: %d", strings.Join(re.rec.Addrs, ","), re.rec.Port)
}

// SetupEndpoint sets the endpoint up and returns the error of a setting it
// cannot use. It logs that error as well, since its callers drop it, and
// keeps it for Send, which returns it on every call: a bad configuration
// then gives an error that names its cause on every REST call.
func SetupEndpoint(rn *RestEndpoint, cfg *jcom.RestEndpointCfg, logger *logrus.Entry) (err error) {
	rn.rec = *cfg

	rn.l = logger.WithFields(logrus.Fields{"section": "rest"})

	rn.l.Debugf("Setup rest endpoint for addresses %v", cfg.Addrs)

	if err = SetupRestProxy(&rn.rp, cfg, rn.l); err != nil {
		rn.rp.httpRestProxy = nil
		rn.rp.setupErr = err
		rn.l.Errorf("cannot set up the REST client for %v: %s", cfg.Addrs, err)
		return err
	}
	return nil
}
