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
	CreateVolume(vdesc CreateVolumeDescriptor) RestError
	GetVolume(vname string) (*Volume, RestError)
	DeleteVolume(vname string, rSnapshots bool) RestError
	ListVolumes() ([]string, RestError)

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

type Storage struct {
	addr string
	port int
	user string
	pass string
	pool string
	rp   RestProxyInterface
	l    *logrus.Entry

	prot string
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

func (s *Storage) String() string {
	return s.addr
}

func NewProvider(cfg *StorageCfg, l *logrus.Entry) (s *Storage, err error) {

	rpc := RestProxyCfg{
		Addr:        cfg.Addr,
		Port:        cfg.Port,
		Prot:        cfg.Prot,
		User:        cfg.User,
		Pass:        cfg.Pass,
		Pool:        cfg.Pool,
		IdleTimeOut: cfg.IdleTimeOut,
		Tries:       cfg.Tries,
	}

	restProxy, err := NewRestProxy(rpc, l)
	if err != nil {
		l.Errorf("cannot create REST client for: %s", cfg.Addr)
	}

	s = &Storage{
		addr: cfg.Addr,
		port: cfg.Port,
		user: cfg.User,
		pass: cfg.Pass,
		pool: cfg.Pool,
		rp:   restProxy,
		l:    l,
	}

	l = l.WithFields(logrus.Fields{
		"obj":     "Storage",
		"storage": s.addr + ":" + string(s.port),
		"pool":    s.pool,
	})

	l.Debugf("Created for %s", cfg.Addr)

	return s, nil
}
