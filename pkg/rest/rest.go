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
	"github.com/sirupsen/logrus"
)

var (
	vendorVersion = "dev"
)

type StorageInterface interface {

	// General information
	GetAddress() (string, int)

	// Pools
	GetPool() (*Pool, RestError)
	GetPools() ([]PoolShort, error)

	// Volumes
	CreateVolume(poolName string, vol Volume) RestError
	GetVolume(vname string) (*Volume, RestError)
	DeleteVolume(vname string, rSnapshots bool) RestError
	ListVolumes(poolName string, vols *[]Volume) (RestError)

	CreateSnapshot(vname string, sname string) RestError
	GetSnapshot(vname string, sname string) (*Snapshot, RestError)
	DeleteSnapshot(vname string, sname string) RestError
	ListAllSnapshots(f func(string) bool) ([]SnapshotShort, RestError)
	ListVolumeSnapshots(string, func(string) bool) ([]SnapshotShort, RestError)

	GetTarget(tname string) (*Target, RestError)
	CreateTarget(tname string) RestError
	DeleteTarget(tname string) RestError

	AttachToTarget(tname string, vname string, mode string) RestError
	DettachFromTarget(tname string, vname string) RestError

	AddUserToTarget(tname string, name string, pass string) RestError

	CreateClone(vname string, sname string, cname string) RestError
	DeleteClone(vname string, sname string, cname string, rChildren bool, rDependent bool) RestError
	PromoteClone(vname string, sname string, cname string) RestError
}


type RestEndpointCfg struct {
	Addrs        []string
	Port        int
	Prot        string
	User        string
	Pass        string
	IdleTimeOut string // See time Duration
	Tries       int
}

type RestEndpoint struct {
	rec  RestEndpointCfg
	rp   RestProxy
	l    *logrus.Entry
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
	return re.rec.Addrs[0]
}

// func GetEndpoint(cfg *RestEndpointCfg, l *logrus.Entry) (s *RestEndpoint, err error) {
	
//	return nil, nil
	// restProxy, err := NewRestProxy(cfg, l)
	// if err != nil {
	// 	l.Errorf("cannot create REST client for: %s", cfg.Addrs[0])
	// }

	// s = &Storage{
	// 	addr: cfg.Addrs[0],
	// 	port: cfg.Port,
	// 	user: cfg.User,
	// 	pass: cfg.Pass,
	// 	rp:   restProxy,
	// 	l:    l,
	// }

	// l = l.WithFields(logrus.Fields{
	// 	"obj":     "RestEndpoint",
	// 	"storage": cfg.Addrs[0] + ":" + string(cfg.Port),
	// })

	// l.Debugf("Created for %s", cfg.Addrs[0])

	// return s, nil
// }

func SetupEndpoint(rn *RestEndpoint, cfg *RestEndpointCfg, logger *logrus.Entry) (err error) {

	rn.rec = *cfg
	
	rn.l = logger.WithFields(logrus.Fields{"section":     "rest",})

	rn.l.Debugf("Setup rest endpoint for addresses %v", cfg.Addrs)

	if ser := SetupRestProxy(&rn.rp, cfg, rn.l); ser != nil {
		logrus.Errorf("cannot create REST client for: %v", cfg.Addrs)
	}
	
	// rn.l.Debugf("RP log value %+s", rn.rp)
	
	// var v []Volume
	// rn.ListVolumes("Pool-0", &v)
	return nil
}
