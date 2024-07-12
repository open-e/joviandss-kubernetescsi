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

func (s *RestEndpoint) CreateNASVolume(ctx context.Context, desc *CreateNASVolumeDescriptor) RestError {
	addr := fmt.Sprintf("api/v3/shares")

	l := jcom.LFC(ctx)
	l = l.WithFields(log.Fields{
		"func":    "CreateNASVolume",
		"section": "rest",
		"url":     addr,
	})

	var share ResourceN
	rsp := GeneralResponse{Data: &share}

	stat, body, err := s.rp.Send(ctx, "POST", addr, desc, GetVolumeRCode)
	if err != nil {
		msg := fmt.Sprintf("Unable to create share %s ", desc.Name)
		l.Warn(msg)
		return GetError(RestErrorRequestMalfunction, msg)
	}

	if errU := s.unmarshal(body, &rsp); errU != nil {
		return errU
	}

	if stat == CodeOK || stat == CodeCreated {
		return nil
	}

	return getError(ctx, body)
}
