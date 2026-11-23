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

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	log "github.com/sirupsen/logrus"
)

// GetNASVolumeSnapshot provides information about specific NAS volume snapshot requested
func (s *RestEndpoint) GetNASVolumeSnapshot(ctx context.Context, pool string, vname string, sname string) (sdp *ResourceNASVolumeSnapshot, err RestError) {
	addr := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s/snapshots/%s", pool, vname, sname)

	l := jcom.LFC(ctx)

	l = s.l.WithFields(log.Fields{
		"section": "rest",
		"func":    "GetNASVolumeSnapshot",
		"url":     addr,
	})

	stat, body, err := s.rp.Send(ctx, "GET", addr, nil, GetSnapshotRCode)
	if err != nil {
		l.Warnf("Unable to get volume snapshots %+v", err.Error())
		return nil, err
	}

	var snapdata ResourceNASVolumeSnapshot
	rsp := &GeneralResponse{Data: &snapdata}

	if stat == CodeOK || stat == CodeAccepted {
		l.Debug("Obtained volume listing")
		if err = s.unmarshal(body, &rsp); err != nil {
			return nil, err
		}
		return &snapdata, nil
	}

	return nil, getError(ctx, body)
}
