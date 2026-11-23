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

func (re *RestEndpoint) String() string {
	var ret string

	if len(re.rec.Addrs) > 0 {
		ret += " addres:"
		for _, val := range re.rec.Addrs {
			ret += val
		}
	}
	ret += fmt.Sprintf(" port: %d", re.rec.Port)

	return re.rec.Addrs[0]
}

func SetupEndpoint(rn *RestEndpoint, cfg *jcom.RestEndpointCfg, logger *logrus.Entry) (err error) {
	rn.rec = *cfg

	rn.l = logger.WithFields(logrus.Fields{"section": "rest"})

	rn.l.Debugf("Setup rest endpoint for addresses %v", cfg.Addrs)

	if ser := SetupRestProxy(&rn.rp, cfg, rn.l); ser != nil {
		logrus.Errorf("cannot create REST client for: %v", cfg.Addrs)
	}
	return nil
}
