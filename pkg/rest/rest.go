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

type StorageInterface interface {
	// General information
	GetAddress() (string, int)

	// Pools
	GetPool() (*ResourcePool, RestError)
	// GetPools() ([]PoolShort, error)

	// Volumes
	CreateVolume(poolName string, vol ResourceVolume) RestError
	GetVolume(vname string) (*ResourceVolume, RestError)
	DeleteVolume(vname string, rSnapshots bool) RestError
	ListVolumes(poolName string, vols *[]ResourceVolume) RestError

	CreateSnapshot(vname string, sname string) RestError
	GetSnapshot(vname string, sname string) (*ResourceSnapshot, RestError)
	DeleteSnapshot(vname string, sname string) RestError
	ListAllSnapshots(f func(string) bool) ([]ResourceSnapshotShort, RestError)
	ListVolumeSnapshots(string, func(string) bool) ([]ResourceSnapshotShort, RestError)

	GetTarget(tname string) (*Target, RestError)
	CreateTarget(tname string) RestError
	DeleteTarget(tname string) RestError

	AttachToTarget(tname string, vname string, mode string) RestError
	DettachFromTarget(tname string, vname string) RestError

	AddUserToTarget(tname string, name string, pass string) RestError

	CreateClone(vname string, sname string, cname string) RestError
	// DeleteClone(vname string, sname string, cname string, rChildren bool, rDependent bool) RestError
	PromoteClone(vname string, sname string, cname string) RestError
}

type RestEndpoint struct {
	rec jcom.RestEndpointCfg
	rp  RestProxy
	l   *logrus.Entry
}

type StorageCfg struct {
	Name string
	Addr string
	Port int
	User string
	Pass string
	Pool string

	Prot        string
	Tries       int
	IdleTimeOut string
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
