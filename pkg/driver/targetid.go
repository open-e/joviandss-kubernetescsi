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
	"crypto/sha256"
	"fmt"

	jrest "joviandss-kubernetescsi/pkg/rest"
)

func TargetIQN(prefix string, ld LunDesc) (*string, jrest.RestError) {
	iqn := fmt.Sprintf("%s:%x", prefix, sha256.Sum256([]byte(ld.VDS())))

	if len(iqn) > 255 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Resulting target name is too long %s", iqn))
	}

	return &iqn, nil
}

