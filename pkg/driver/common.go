/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
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

package driver

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

func getResourcesList[RestResource any](ctx context.Context, maxret int, token CSIListingToken,
	grf func(context.Context, CSIListingToken) ([]RestResource, jrest.RestError),
	BasedID func(RestResource) string,
) (lres []RestResource, nt *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "getResourceList",
		"section": "driver",
	})

	l.Debugf("Processing %T", lres)

	for {
		// l.Debugf("lres at start len %d lres %+v", len(lres), lres)

		if ent, err := grf(ctx, token); err != nil {
			return nil, nil, err
		} else {
			// No new snapshots, return what we have so far
			// data, ok := ent.Entries.()

			if len(ent) == 0 {
				return lres, nil, nil
			}

			if len(token.BasedID()) == 0 {
				lres = append(lres, (ent)...)
			} else if token.BasedID() <= BasedID(ent[0]) {
				lres = append(lres, (ent)...)
				token.DropBasedID()
			} else {
				for i, e := range ent {
					if BasedID(e) == token.BasedID() {
						lres = append(lres, ent[i:]...)
						token.BasedID()
						break
					}
				}
			}
		}

		if maxret > 0 {
			if len(lres) >= maxret {
				// l.Debugf("lres at max len %d lres %+v", len(lres), lres)

				if newToken, err := NewCSIListingTokenFromBasedID(BasedID(lres[maxret-1]), token.Page(), token.DC()); err != nil {
					return nil, nil, err
				} else {
					lres = lres[:maxret]
					return lres, &newToken, nil
				}
			}
		}
		token.PageUp()
	}
}
