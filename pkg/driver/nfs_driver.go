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
	// "crypto/sha256"
	"fmt"
	"strings"
	// "time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

const (
	minSupportedVolumeSize = 16 * mib
)

// JovianDSS CSI plugin
type CSINFSDriver struct {
	re jrest.RestEndpoint
	l  *logrus.Entry
}

func (d *CSINFSDriver) cloneLUN(ctx context.Context, pool string, source LunDesc, clone LunDesc, snap *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "cloneLUN",
	})

	l.Debugf("Start cloning")

	var sds string
	if snap == nil {
		snapdata := jrest.CreateNASSnapshotDescriptor{Name: clone.VDS()}

		if err := d.re.CreateNASSnapshot(ctx, pool, source.VDS(), &snapdata); err != nil {
			code := err.GetCode()
			if code != jrest.RestErrorResourceExists {
				return err
			}
			// TODO: check if specific snapshot have clones, if it does and name of clone
			//	is the same as name of the clone, then return error
			//	if there are no clones snapshot have to be deleteted and recreated
		}
		sds = clone.VDS()
	} else {
		sds = snap.sds
	}

	clonedata := jrest.CloneVolumeDescriptor{Name: clone.VDS(), Snapshot: sds}
	err := d.re.CreateClone(ctx, pool, source.VDS(), clonedata)
	return err
}

func (d *CSINFSDriver) CreateVolume(ctx context.Context, pool string, nvd *VolumeDesc, requiredBytes int64, limitBytes int64) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateVolume",
		"section": "NFSDriver",
	})

	ctx = jcom.WithLogger(ctx, l)

	name := nvd.VDS()

	nasvolume := jrest.CreateNASVolumeDescriptor{
		Name: name,
	}

	if requiredBytes > minSupportedVolumeSize {
		rbs := fmt.Sprintf("%d", requiredBytes)
		nasvolume.Reservation = &rbs
	}

	if limitBytes >= minSupportedVolumeSize {
		lbs := fmt.Sprintf("%d", limitBytes)
		nasvolume.Quota = &lbs
	}

	err := d.re.CreateNASVolume(ctx, pool, &nasvolume)

	switch jrest.ErrCode(err) {
	case jrest.RestErrorOk:
		break
	case jrest.RestErrorResourceExists:
		l.Warnf("NAS Volume %s exists", name)
		break
	default:
		l.Debugf("Unable to create nas volume %s, error had happaned %+v", nvd.Name(), err.Error())
		return err
	}

	return d.createShare(ctx, pool, name)
}

// createShare makes the NFS share of the NAS volume name, with the path
// <pool>/<name>/<name>. When the appliance answers that a share of that
// name exists, the share is read, and it is taken as this volume's own only
// when its path is the one sent: a repeated create then succeeds. A share
// with another path fails the create.
func (d *CSINFSDriver) createShare(ctx context.Context, pool string, name string) jrest.RestError {
	enabled, allSquash, active := true, true, true
	path := fmt.Sprintf("%s/%s/%s", pool, name, name)

	desc := jrest.CreateShareDescriptor{
		Name:   name,
		Active: &active,
		Path:   path,
		NFS:    &jrest.ShareNFSDescriptor{Enabled: &enabled, AllSquash: &allSquash},
	}

	err := d.re.CreateShare(ctx, &desc)
	if jrest.ErrCode(err) != jrest.RestErrorResourceExists {
		return err
	}

	share, rerr := d.re.GetShare(ctx, name)
	if rerr != nil {
		return rerr
	}
	if share.Path != path {
		return jrest.GetError(jrest.RestErrorResourceExists,
			fmt.Sprintf("share %s exists with the path %s, and the volume needs %s", name, share.Path, path))
	}
	return nil
}

func (d *CSINFSDriver) CreateVolumeFromSnapshot(ctx context.Context, pool string, sd *SnapshotDesc, nvd *VolumeDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateVolumeFromSnapshot",
		"section": "driver",
	})

	clonedesk := jrest.CloneNASVolumeDescriptor{Name: nvd.VDS()}
	// TODO: implement create from clone

	err := d.re.CreateNASClone(ctx, pool, sd.ld.VDS(), sd.SDS(), clonedesk)

	switch jrest.ErrCode(err) {
	case jrest.RestErrorOk:
		break
	case jrest.RestErrorResourceExists:
		l.Warnf("NAS Volume %s exists", sd.ld.VDS())
		break
	default:
		l.Debugf("Unable to create nas volume %s, error had happaned %+v", nvd.Name(), err.Error())
		return err
	}
	return d.createShare(ctx, pool, nvd.VDS())
}

func (d *CSINFSDriver) deleteIntermediateSnapshot(ctx context.Context, pool string, vds string, sds string) jrest.RestError {
	// The appliance does not refuse the delete of a NAS snapshot that has a
	// clone: it deletes the snapshot and the clone with it. So the clones
	// are read first, and the snapshot is deleted only when the appliance
	// answers that it has none.
	deleted, clones, err := deleteSnapshotWithoutClones(ctx,
		func(ctx context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError) {
			return d.re.GetNASVolumeSnapshotClones(ctx, pool, vds, sds)
		},
		func(ctx context.Context) jrest.RestError {
			return d.re.DeleteNASSnapshot(ctx, pool, vds, sds)
		})
	switch {
	case err != nil || deleted:
		return err
	case len(clones) == 1 && clones[0].Name == sds:
		// the volume this snapshot was made for exists already
		return jrest.GetError(jrest.RestErrorResourceExists,
			fmt.Sprintf("NAS volume %s made from volume %s already exists", sds, vds))
	case len(clones) == 1:
		return jrest.GetError(jrest.RestErrorResourceBusySnapshotHasClones,
			fmt.Sprintf("intermediate snapshot %s of volume %s has the clone %s", sds, vds, clones[0].Name))
	default:
		return jrest.GetError(jrest.RestErrorStorageFailureUnknown,
			fmt.Sprintf("intermediate snapshot %s of volume %s has %d clones", sds, vds, len(clones)))
	}
}

// CreateVolumeFromVolume creates volume from another volume
//
//	Takes as arguments:
//
//	- ctx context
//	- pool pool name
//	- vd source volume descripto
//	- nvd new volume desctiptor
func (d *CSINFSDriver) CreateVolumeFromVolume(ctx context.Context, pool string, vd *VolumeDesc, nvd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateVolumeFromVolume",
		"section": "driver",
	})

	// return jrest.GetError(jrest.RestErrorFailureUnknown,
	//	fmt.Sprintf("Creation of volume from volume is not implemented"))
	snapdata := jrest.CreateNASSnapshotDescriptor{Name: nvd.VDS()}

	if err := d.re.CreateNASSnapshot(ctx, pool, vd.VDS(), &snapdata); err != nil {
		code := err.GetCode()
		// We are not able to create this snapshot for some reason

		// Probably it was already created
		if code == jrest.RestErrorResourceExists {
			d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS())
		}
		return err
	}

	clonedata := jrest.CloneNASVolumeDescriptor{Name: nvd.VDS()}
	if err = d.re.CreateNASClone(ctx, pool, vd.VDS(), nvd.VDS(), clonedata); err != nil {
		l.Warnf("Unable to create NAS volume %s from snapshot %s of volume %s, because of error %+v. Removing intermediate snapshot", nvd.VDS(), nvd.VDS(), vd.VDS(), err.Error())

		d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS())
		return err
	}

	return d.createShare(ctx, pool, nvd.VDS())
}

// cleanIntermediateSnapshots request list of snapshots related to particular volume and check if there are intermediate one that can be deleted
//
//	return list of snapshots that does not contain 'handing' one or error
func (d *CSINFSDriver) cleanIntermediateSnapshots(ctx context.Context, pool string, vd *VolumeDesc) (out []keptSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "cleanIntermediateSnapshots",
		"section": "driver",
	})
	snaps, gserr := d.volumeSnapshots(ctx, pool, vd)
	if gserr != nil {
		l.Debugf("Unable to get list of snapshots for volume %s", vd.Name())
		return nil, gserr
	}
	l.Debugf("Clean intermediate snapshots return %d records", len(snaps))

	for _, snap := range snaps {
		if isIntermediate(snap.Name) {
			// An intermediate snapshot. The appliance deletes a NAS snapshot
			// that has a clone, and the clone with it, and a list entry does
			// not say whether the snapshot has clones. So the appliance is
			// asked, and the snapshot is deleted only when it answers that
			// there is none.
			sname := snap.Name
			deleted, clones, derr := deleteSnapshotWithoutClones(ctx,
				func(ctx context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError) {
					return d.re.GetNASVolumeSnapshotClones(ctx, pool, vd.VDS(), sname)
				},
				func(ctx context.Context) jrest.RestError {
					return d.re.DeleteNASSnapshot(ctx, pool, vd.VDS(), sname)
				})
			if derr != nil {
				return nil, derr
			}
			if !deleted {
				out = append(out, keptSnapshot{Name: sname, Clones: clonesNames(clones)})
			}
		} else {
			// a CSI snapshot, or a snapshot the plugin did not make: it stays
			out = append(out, keptSnapshot{Name: snap.Name})
		}
	}

	return out, nil
}

func (d *CSINFSDriver) deleteNASVolAndShare(ctx context.Context, pool string, vd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteLUN",
		"section": "driver",
	})

	// err = d.re.DeleteShare(ctx, vd.VDS(), nil)
	// switch jrest.ErrCode(err) {
	// case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEShare:
	// 	break
	// case jrest.RestErrorOk:
	// 	break
	// default:
	// 	l.Debugf("Unable to delete share %s, error had happaned %+v", vd.Name(), err.Error())
	// 	return err
	// }

	// forceUmountNASVolume := true
	// delnas := jrest.DeleteNASVolumeDescriptor{ForceUmount: &forceUmountNASVolume}

	err = d.re.DeleteNASVolume(ctx, pool, vd.VDS())
	switch jrest.ErrCode(err) {
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEVolume:
		l.Debugf("NAS volume %s do not exists. Do nothing.", vd.Name())
		break
	case jrest.RestErrorOk:
		break
	default:
		l.Debugf("Unable to delete nas volume %s, error had happaned %+v", vd.Name(), err.Error())
		return err
	}

	return nil
}

// deleteLUN deletes a NAS volume. The appliance does not refuse the delete
// of a NAS volume that has a snapshot: it deletes the snapshots with it. So
// the driver looks first: it reads the snapshots of the volume, runs the
// cleanup of the intermediate ones, and sends the delete only when no
// snapshot is left. A volume that still has snapshots is not deleted: the
// error names them, and its code is the one the controller answers
// FailedPrecondition.
//
// A snapshot made between the read and the delete is not seen, and goes
// with the volume: the appliance has no delete that spares snapshots.
func (d *CSINFSDriver) deleteLUN(ctx context.Context, pool string, vd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteLUN",
		"section": "driver",
	})

	snaps, gserr := d.cleanIntermediateSnapshots(ctx, pool, vd)
	switch jrest.ErrCode(gserr) {
	case jrest.RestErrorOk:
		if len(snaps) > 0 {
			// the volume has snapshots, and stays
			l.Debugf("Snapshots after cleaning intermediate one %+v", snaps)
			return volumeHasSnapshots(vd.Name(), snaps)
		}
	case jrest.RestErrorResourceDNE:
		// The volume is not there, so it has no snapshot to lose. The delete
		// is sent all the same, and answers for itself.
		l.Debugf("NAS volume %s has no snapshot list: %s", vd.Name(), gserr.Error())
	default:
		// the snapshots could not be read: nothing is deleted
		return gserr
	}

	err = d.deleteNASVolAndShare(ctx, pool, vd)

	switch jrest.ErrCode(err) {
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEVolume:
		l.Debugf("NAS volume %s do not exists. Do nothing.", vd.Name())
		return nil
	case jrest.RestErrorOk:
		return nil
	default:
		// RestErrorResourceBusy among them: a volume whose snapshot has a
		// clone outside the plugin's names
		l.Debugf("Unable to delete lun %s, error had happaned %+v", vd.Name(), err.Error())
		return err
	}
}

func (d *CSINFSDriver) DeleteVolume(ctx context.Context, pool string, vid *VolumeDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "DeleteVolume",
		"section": "driver",
	})
	l.Debugf("Deleting volume %s", vid.Name())
	return d.deleteLUN(ctx, pool, vid)
}

// ListAllVolumes returns the NAS volumes of the plugin on the pool that
// follow the place the token gives, maxret of them at most, all when maxret
// is 0, and the token of the next call when volumes follow.
func (d *CSINFSDriver) ListAllVolumes(ctx context.Context, pool string, maxret int, token *CSIListingToken) (vols interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllVolumes",
		"proto":   "NFS",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceNASVolume, jrest.RestError) {
			return d.re.GetNASVolumesEntries(ctx, pool, page, perPage)
		},
		func(v jrest.ResourceNASVolume) string { return v.Name },
		sameNASVolume,
		func(v jrest.ResourceNASVolume) bool { return ownVolumeName(l, v.Name) },
		func(v jrest.ResourceNASVolume) string { return keyPart(v.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// ownSnapshot reports whether an entry of a snapshot list is a snapshot of
// a NAS volume. Both lists of all snapshots return the snapshots of zvols
// and of NAS volumes together, so this driver drops those of zvols; another
// kind of parent is kept.
func (d *CSINFSDriver) ownSnapshot(s jrest.ResourceNASSnapshotShort) bool {
	return s.Parent.Type != jrest.SnapshotParentVolume
}

// ListAllSnapshots does for the NAS snapshots of the pool what
// ListAllVolumes does for its NAS volumes, in the order of the parent's
// name and then of the snapshot's.
func (d *CSINFSDriver) ListAllSnapshots(ctx context.Context, pool string, maxret int, token *CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllSnapshots",
		"proto":   "NFS",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceNASSnapshotShort, jrest.RestError) {
			return d.re.GetNASSnapshotsEntries(ctx, pool, page, perPage)
		},
		func(s jrest.ResourceNASSnapshotShort) string { return s.Name },
		sameNASSnapshot,
		func(s jrest.ResourceNASSnapshotShort) bool {
			return d.ownSnapshot(s) && ownSnapshotNames(l, s.Parent.Name, s.Name)
		},
		func(s jrest.ResourceNASSnapshotShort) string { return snapshotsListKey(s.Parent.Name, s.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// ListVolumeSnapshots provides maxret records of snapshots of volume starting from token
// if no token nor limit on number of snapshot is given it will list all snapshots of particular volume
func (d *CSINFSDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeDesc, maxret int, token *CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListNASVolumeSnapshots",
		"proto":   "NFS",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceNASSnapshotShort, jrest.RestError) {
			return d.re.GetNASVolumeSnapshotsEntries(ctx, pool, vid.VDS(), page, perPage)
		},
		func(s jrest.ResourceNASSnapshotShort) string { return s.Name },
		sameNASSnapshot,
		func(s jrest.ResourceNASSnapshotShort) bool {
			return d.ownSnapshot(s) && ownSnapshotName(l, vid, s.Name)
		},
		func(s jrest.ResourceNASSnapshotShort) string { return keyPart(s.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// volumeSnapshots reads every snapshot of one NAS volume, with no token and
// no limit, and whatever its name: the cleanup looks at the snapshots the
// plugin did not make too.
func (d *CSINFSDriver) volumeSnapshots(ctx context.Context, pool string, vd *VolumeDesc) ([]jrest.ResourceNASSnapshotShort, jrest.RestError) {
	read, err := readList(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceNASSnapshotShort, jrest.RestError) {
			return d.re.GetNASVolumeSnapshotsEntries(ctx, pool, vd.VDS(), page, perPage)
		})
	if err != nil {
		return nil, err
	}
	var snaps []jrest.ResourceNASSnapshotShort
	for _, s := range mergePages(read, func(s jrest.ResourceNASSnapshotShort) string { return s.Name }, sameNASSnapshot) {
		if d.ownSnapshot(s) {
			snaps = append(snaps, s)
		}
	}
	return snaps, nil
}

func NewJovianDSSCSINFSDriver(cfg *jcom.RestEndpointCfg, l *logrus.Entry) (d CSIDriver, err error) {
	var drvr CSINFSDriver
	jrest.SetupEndpoint(&drvr.re, cfg, l)

	return &drvr, nil
}

func (d *CSINFSDriver) GetVolume(ctx context.Context, pool string, vd *VolumeDesc) (out *GeneralVolume, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetVolume",
		"section": "driver",
	})

	l.Debugf("Get volume with id: %s", vd.VDS())

	if nv, err := d.re.GetNASVolume(ctx, pool, vd.VDS()); err != nil {
		return nil, err
	} else {
		gv := GeneralVolume{
			Name:      nv.Name,
			Size:      nv.GetSize(),
			OVolume:   nv.OriginVolume(),
			OSnapshot: nv.OriginSnapshot(),
		}
		return &gv, nil
	}
	// v for Volume
}

func (d *CSINFSDriver) GetSnapshot(ctx context.Context, pool string, vd LunDesc, sd *SnapshotDesc) (out *GeneralSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetSnapshot",
		"section": "driver",
	})

	l.Debugf("Get snapshot %s of volume %s", sd.SDS(), vd.VDS())

	if nv, err := d.re.GetNASVolumeSnapshot(ctx, pool, vd.VDS(), sd.SDS()); err != nil {
		return nil, err
	} else {
		gv := GeneralSnapshot{
			Name:     nv.Name,
			VolSize:  nv.GetSize(),
			Creation: nv.GetCreation(),
		}
		return &gv, nil
	}
}

func (d *CSINFSDriver) CreateSnapshot(ctx context.Context, pool string, vd *VolumeDesc, sd *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateSnapshot",
		"section": "driver",
	})

	l.Debugf("Create snapshot %s for volume %s", sd.SDS(), vd.VDS())

	snapdata := jrest.CreateNASSnapshotDescriptor{Name: sd.SDS()}

	return d.re.CreateNASSnapshot(ctx, pool, vd.VDS(), &snapdata)
}

func (d *CSINFSDriver) DeleteSnapshot(ctx context.Context, pool string, ld LunDesc, sd *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "DeleteSnapshot",
		"section": "driver",
	})

	l.Debugf("Delete snapshot %s for volume %s", sd.SDS(), ld.VDS())

	// forceUmount := true
	// deldata := jrest.DeleteNASSnapshotDescriptor{ForceUmount: &forceUmount}

	err := d.re.DeleteNASSnapshot(ctx, pool, ld.VDS(), sd.SDS())

	var dvols []string
	var dsnaps []string
	var ncsi []string
	var msg string

	switch jrest.ErrCode(err) {
	case jrest.RestErrorOk:
		l.Debugf("Snapshot %s deleted", sd.Name())
		return nil
	case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusySnapshotHasClones:
		if clones, rErr := d.re.GetNASVolumeSnapshotClones(ctx, pool, ld.VDS(), sd.SDS()); rErr != nil {
			return rErr
		} else {
			for _, clone := range clones {
				if IsSDS(clone.Name) {
					dsnaps = append(dsnaps, clone.Name)
				} else if IsVDS(clone.Name) {
					dvols = append(dvols, clone.Name)
				} else {
					ncsi = append(ncsi, clone.Name)
				}
			}
		}

	case jrest.RestErrorResourceDNE:
		l.Debugf("snapshot %s do not exists", sd.SDS())
		return nil
	default:
		return err
	}

	if len(dsnaps) > 0 {
		// forceUmount := true
		// delclone := jrest.DeleteVolumeDescriptor{ForceUmount: &forceUmount}

		for _, snapclone := range dsnaps {
			if rErr := d.re.DeleteNASClone(ctx, pool, ld.VDS(), sd.SDS(), snapclone); rErr != nil {
				msg = fmt.Sprintf("Unable to delete snapshot %s with ID %s because it has volume associated with it %s that cant be deleted, please delete physical zvol first", sd.Name(), sd.CSIID(), snapclone)
				return jrest.GetError(jrest.RestErrorResourceBusy, msg)
			}
		}
	}

	msg = fmt.Sprintf("Snapshot %s with ID %s is dependent upon by", sd.Name(), sd.CSIID())

	if len(dvols) > 0 {
		msg += fmt.Sprintf(" clones: %s", strings.Join(dvols[:], ","))
	}
	if len(ncsi) > 0 {
		msg += fmt.Sprintf(" not CSI related clones: %s", strings.Join(ncsi[:], ","))
	}

	err = jrest.GetError(jrest.RestErrorResourceBusy, msg)
	return err
}

func (d *CSINFSDriver) GetPool(ctx context.Context, pool string) (out *jrest.ResourcePool, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetPool",
		"section": "driver",
	})

	l.Debugf("Get pool with id: %s", pool)

	return d.re.GetPool(ctx, pool)
}

func (d *CSINFSDriver) PublishVolume(ctx context.Context, pool string, ld LunDesc, iqnPrefix string, readonly bool) (iscsiContext *map[string]string, rErr jrest.RestError) {
	// Create target

	iContext := map[string]string{}

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "PublishVolume",
		"proto":   "NFS",
		"section": "driver",
	})

	// Attach to target
	var mode string = "wt"
	if readonly == true {
		mode = "ro"
	}

	// var attachLun jrest.TargetLunDescriptor

	// attachLun.Name = ld.VDS()
	// attachLun.Mode = &mode
	// lunID := 0
	// attachLun.LUN = &lunID

	// rErr = d.re.AttachVolumeToTarget(ctx, pool, iqn, &attachLun)

	// if rErr != nil {
	// 	code := rErr.GetCode()
	// 	switch code {
	// 	case jrest.RestErrorResourceDNEVolume:
	// 		d.re.DeleteTarget(ctx, pool, tname)
	// 		return nil, rErr
	// 	case jrest.RestErrorResourceBusy:
	// 		// According to specification from
	// 		// TODO: check that resource indeed properly assigned and continue if everything is ok
	// 		l.Debugf("Volume %s already attached", ld.Name())
	// 	default:
	// 		return nil, rErr
	// 	}
	// }
	iContext["protocol_type"] = "NFS"
	iContext["read_only"] = mode
	iContext["share_path"] = fmt.Sprintf("/%s", ld.VDS())

	// iContext["target"] = tname
	// iContext["lun"] = fmt.Sprintf("%d", lunID)
	return &iContext, nil

	// for i := 0; i < 3; i++ {
	// 	target, rErr := d.re.GetTarget(ctx, pool, iqn)
	// 	switch jrest.ErrCode(rErr) {
	// 	case jrest.RestErrorOk:
	// 		if target.Active == true {
	// 			return &iContext, nil
	// 		}
	// 	case jrest.RestErrorResourceDNE:
	// 		// According to specification from
	// 		time.Sleep(time.Second)
	// 		continue
	// 	default:
	// 		continue
	// 	}
	// }

	// return nil, jrest.GetError(jrest.RestErrorRequestTimeout, fmt.Sprintf("Unable to ensure that target %s is up and running", iqn))
}

func (d *CSINFSDriver) UnpublishVolume(ctx context.Context, pool string, prefix string, ld LunDesc) (rErr jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "UnpublishVolume",
		"proto":   "NFS",
		"section": "driver",
	})

	// We want target name to be uniquee
	// iqn, rErr := TargetIQN(prefix, ld)

	// if rErr != nil {
	// 	return rErr
	// }

	// rErr = d.re.DettachVolumeFromTarget(ctx, pool, *iqn, ld.VDS())

	// if rErr != nil {
	// 	code := rErr.GetCode()
	// 	switch code {
	// 	case jrest.RestErrorOk:
	// 		l.Debugf("Volume %s was detached from target %s", ld.Name(), *iqn)
	// 	case jrest.RestErrorResourceDNEVolume, jrest.RestErrorResourceDNE:
	// 		l.Debugf("Volume %s is not attached from target %s", ld.Name(), *iqn)
	// 	case jrest.RestErrorResourceDNETarget:
	// 		return nil
	// 	default:
	// 		return rErr
	// 	}
	// }

	// rErr = d.re.DeleteTarget(ctx, pool, *iqn)

	// switch jrest.ErrCode(rErr) {
	// case jrest.RestErrorOk:
	// 	l.Debugf("target %s deleted", *iqn)
	// case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNETarget:
	// 	l.Debugf("target %s do not exists", *iqn)
	// default:
	// 	return rErr
	// }

	return nil
}
