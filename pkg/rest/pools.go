/*
Copyright (c) 2024 Open-E, Inc.
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
	"fmt"

	log "github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

func (s *RestEndpoint) GetPool(ctx context.Context, pool string) (*ResourcePool, RestError) {

	var respool ResourcePool
	var rsp = GeneralResponse{Data: &respool}

	addr := fmt.Sprintf("api/v3/pools/%s", pool)

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "GetPool",
		"url":     addr,
		"section": "rest",
	})

	l.Debugf("getting pool %s ", pool)

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetPoolRCode)

	if err != nil {
		s.l.Warnln("unable to get pool: ", pool)
		return nil, err
	}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return nil, errU
	}

	if stat == CodeOK || stat == CodeNoContent {
		return &respool, nil
	}
	return nil, getError(ctx, body)
}
