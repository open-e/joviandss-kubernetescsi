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
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// JovianDSS CSI plugin
type CSIISCSiDriver struct {
	re jrest.RestEndpoint
	l  *logrus.Entry
}

func (d *CSIISCSiDriver) cloneLUN(ctx context.Context, pool string, source LunDesc, clone LunDesc, snap *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"proto":   "iSCSI",
		"func":    "cloneLUN",
		"section": "driver",
	})

	l.Debugf("Start cloning")

	var sds string
	if snap == nil {
		snapdata := jrest.CreateSnapshotDescriptor{SnapshotName: clone.VDS()}

		if err := d.re.CreateSnapshot(ctx, pool, source.VDS(), &snapdata); err != nil {
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

func (d *CSIISCSiDriver) CreateVolume(ctx context.Context, pool string, nvd *VolumeDesc, requiredBytes int64, limitBytes int64) jrest.RestError {
	var volumeSize int64
	if limitBytes == requiredBytes {
		volumeSize = limitBytes
	} else if requiredBytes < minSupportedVolumeSize {
		volumeSize = minSupportedVolumeSize
	} else {
		volumeSize = requiredBytes
	}

	vd := jrest.CreateVolumeDescriptor{
		Name: nvd.VDS(),
		Size: fmt.Sprintf("%d", volumeSize),
	}

	return d.re.CreateVolume(ctx, pool, vd)
}

func (d *CSIISCSiDriver) CreateVolumeFromSnapshot(ctx context.Context, pool string, sd *SnapshotDesc, nvd *VolumeDesc) jrest.RestError {
	clonedata := jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: sd.SDS()}
	return d.re.CreateClone(ctx, pool, sd.ld.VDS(), clonedata)
}

func (d *CSIISCSiDriver) deleteIntermediateSnapshot(ctx context.Context, pool string, vds string, sds string) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteIntermediateSnapshot",
		"proto":   "iSCSI",
		"section": "driver",
	})
	// force_umount false: a delete never detaches a zvol from its target;
	// only DettachVolumeFromTarget does that.
	forceUmount := false
	snapdeldata := jrest.DeleteSnapshotDescriptor{ForceUmount: &forceUmount}
	// Just in case lets delete this snapshot and do everything from groud up
	if err = d.re.DeleteSnapshot(ctx, pool, vds, sds, snapdeldata); err != nil {
		code := err.GetCode()
		// Removing this snapshot is not possible

		// May be somebody is using this snapshot already to make volumes from it
		if code == jrest.RestErrorResourceBusySnapshotHasClones {
			// We gona check if volume name is exactly volume that we have to create before returning success
			if snap, errGS := d.re.GetVolumeSnapshot(ctx, pool, vds, sds); errGS != nil {
				// That is a weird, previously we failed because snapshot existed and now it is gone
				// looks like some king of race condition
				if errGS.GetCode() == jrest.RestErrorResourceDNE {
					return jrest.GetError(jrest.RestErrorStorageFailureUnknown,
						fmt.Sprintf("It looks like there is a race condition"))
				}
				return jrest.GetError(jrest.RestErrorStorageFailureUnknown,
					fmt.Sprintf("Unable to delete intermediate snapshot %s for volume %s as clean up operation because of %s",
						vds, sds, errGS.Error()))
			} else {
				// Target volume already created
				if clones := snap.ClonesNames(); len(clones) == 1 {
					if clones[0] == sds {
						return jrest.GetError(jrest.RestErrorResourceExists, fmt.Sprintf("Volume %s created from volume %s already exists", sds, vds))
					}
				} else {
					return jrest.GetError(jrest.RestErrorStorageFailureUnknown, fmt.Sprintf("Intermediate snapshot have multiple clones: %+v, that should never happen", clones))
				}
			}
		} else {
			// some unexpected error
			return jrest.GetError(
				jrest.RestErrorStorageFailureUnknown,
				fmt.Sprintf("Unablet to clear intermediate snapshot %s, please delete it manualy for volume %s. Error %+v ", sds, vds, err.Error()))
		}
	}
	return err
}

// CreateVolumeFromVolume creates volume from another volume
//
//	Takes as arguments:
//
//	- ctx context
//	- pool pool name
//	- vd source volume descripto
//	- nvd new volume desctiptor
func (d *CSIISCSiDriver) CreateVolumeFromVolume(ctx context.Context, pool string, vd *VolumeDesc, nvd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateVolumeFromVolume",
		"proto":   "iSCSI",
		"section": "driver",
	})

	snapdata := jrest.CreateSnapshotDescriptor{SnapshotName: nvd.VDS()}

	if err := d.re.CreateSnapshot(ctx, pool, vd.VDS(), &snapdata); err != nil {
		code := err.GetCode()
		// We are not able to create this snapshot for some reason

		// Probably it was already created
		if code == jrest.RestErrorResourceExists {
			d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS())
		}
		return err
	}

	clonedata := jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: nvd.VDS()}
	if err = d.re.CreateClone(ctx, pool, vd.VDS(), clonedata); err != nil {
		l.Warnf("Unable to create volume %s from snapshot %s of volume %s, because of error %+v. Removing intermediate snapshot", nvd.VDS(), nvd.VDS(), vd.VDS(), err.Error())

		d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS())
		return err
	}

	return nil
}

// cleanIntermediateSnapshots request list of snapshots related to particular volume and check if there are intermediate one that can be deleted
//
//	return list of snapshots that does not contain 'handing' one or error
func (d *CSIISCSiDriver) cleanIntermediateSnapshots(ctx context.Context, pool string, vd *VolumeDesc) (out []keptSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "cleanIntermediateSnapshots",
		"proto":   "iSCSI",
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
			// An intermediate snapshot. A list entry does not say whether
			// the snapshot has clones, so the appliance is asked, and the
			// snapshot is deleted only when it answers that there is none.
			sname := snap.Name
			deleted, clones, derr := deleteSnapshotWithoutClones(ctx,
				func(ctx context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError) {
					return d.re.GetVolumeSnapshotClones(ctx, pool, vd.VDS(), sname)
				},
				func(ctx context.Context) jrest.RestError {
					// force_umount false: a delete never detaches a zvol from its target;
					// only DettachVolumeFromTarget does that.
					forceUmount := false
					snapdeldata := jrest.DeleteSnapshotDescriptor{ForceUmount: &forceUmount}
					return d.re.DeleteSnapshot(ctx, pool, vd.VDS(), sname, snapdeldata)
				})
			if derr != nil {
				if derr.GetCode() == jrest.RestErrorResourceDNE {
					// The snapshot went away after the list was read: it is
					// in the way no longer. Only "not found" for the list
					// itself says that the volume is gone; passed on from
					// here, it would let the volume go with the snapshots
					// that are still to be looked at.
					l.Debugf("Intermediate snapshot %s of volume %s is gone: %s", sname, vd.Name(), derr.Error())
					continue
				}
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

// deleteLUN deletes a volume. The appliance refuses the delete of a zvol
// that has a snapshot, "volume has children", or a clone. Either refusal
// can come from an intermediate snapshot that a clone no longer needs, so
// the cleanup of intermediate snapshots runs for both, and the delete is
// sent again when the cleanup leaves no snapshot. A volume that still has
// snapshots is not deleted: the error names them, and its code is the one
// the controller answers FailedPrecondition.
func (d *CSIISCSiDriver) deleteLUN(ctx context.Context, pool string, vd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteLUN",
		"proto":   "iSCSI",
		"section": "driver",
	})

	// force_umount false: a delete never detaches a zvol from its target;
	// only DettachVolumeFromTarget does that.
	forceUmount := false
	deldata := jrest.DeleteVolumeDescriptor{ForceUmount: &forceUmount}
	err = d.re.DeleteVolume(ctx, pool, vd.VDS(), deldata)

	switch jrest.ErrCode(err) {
	case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusyVolumeHasSnapshots:
		break
	case jrest.RestErrorResourceDNE:
		return nil
	case jrest.RestErrorOk:
		return nil
	default:
		l.Debugf("Unable to delete lun %s, error had happaned %+v", vd.Name(), err.Error())
		return err
	}
	l.Debugf("Volume %s is busy, indentifying relaying resources", vd.Name())

	snaps, gserr := d.cleanIntermediateSnapshots(ctx, pool, vd)
	if gserr != nil {
		switch gserr.GetCode() {
		case jrest.RestErrorResourceDNE:
			return gserr
		default:
			// TODO: think about providing better error information for cases
			// when we are not able to provide proper list of dependent snapshots
			return err
		}
	}
	l.Debugf("Snapshots after cleaning intermediate one %+v", snaps)

	if len(snaps) > 0 {
		// the volume has snapshots, and stays
		return volumeHasSnapshots(vd.Name(), snaps)
	}

	// The cleanup left no snapshot: what refused the delete is gone, or
	// was never a snapshot of this volume.
	err = d.re.DeleteVolume(ctx, pool, vd.VDS(), deldata)
	switch jrest.ErrCode(err) {
	case jrest.RestErrorOk, jrest.RestErrorResourceDNE:
		return nil
	case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusyVolumeHasSnapshots:
		// still refused, with no snapshot to name: the appliance's own words
		return jrest.GetError(jrest.RestErrorResourceBusy,
			fmt.Sprintf("Volume %s is dependent upon by: %s", vd.Name(), err.Error()))
	default:
		return err
	}
}

func (d *CSIISCSiDriver) DeleteVolume(ctx context.Context, pool string, vid *VolumeDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "DeleteVolume",
		"proto":   "iSCSI",
		"section": "driver",
	})

	return d.deleteLUN(ctx, pool, vid)
}

// ListAllVolumes returns the volumes of the plugin on the pool that follow
// the place the token gives, maxret of them at most, all when maxret is 0,
// and the token of the next call when volumes follow.
func (d *CSIISCSiDriver) ListAllVolumes(ctx context.Context, pool string, maxret int, token *CSIListingToken) (vols interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllVolumes",
		"proto":   "iSCSI",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceVolume, jrest.RestError) {
			return d.re.GetVolumesEntries(ctx, pool, page, perPage)
		},
		func(v jrest.ResourceVolume) string { return v.Name },
		sameVolume,
		func(v jrest.ResourceVolume) bool { return ownVolumeName(l, v.Name) },
		func(v jrest.ResourceVolume) string { return keyPart(v.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// ownSnapshot reports whether an entry of a snapshot list is a snapshot of
// a zvol. Both lists of all snapshots return the snapshots of zvols and of
// NAS volumes together, so this driver drops those of NAS volumes; another
// kind of parent is kept.
func (d *CSIISCSiDriver) ownSnapshot(s jrest.ResourceSnapshotShort) bool {
	return s.Parent.Type != jrest.SnapshotParentNASVolume
}

// ListAllSnapshots does for the snapshots of the pool what ListAllVolumes
// does for its volumes, in the order of the parent's name and then of the
// snapshot's.
func (d *CSIISCSiDriver) ListAllSnapshots(ctx context.Context, pool string, maxret int, token *CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllSnapshots",
		"proto":   "iSCSI",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceSnapshotShort, jrest.RestError) {
			return d.re.GetSnapshotsEntries(ctx, pool, page, perPage)
		},
		func(s jrest.ResourceSnapshotShort) string { return s.Name },
		sameSnapshot,
		func(s jrest.ResourceSnapshotShort) bool {
			return d.ownSnapshot(s) && ownSnapshotNames(l, s.Parent.Name, s.Name)
		},
		func(s jrest.ResourceSnapshotShort) string { return snapshotsListKey(s.Parent.Name, s.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// ListVolumeSnapshots provides maxret records of snapshots of volume starting from token
// if no token nor limit on number of snapshot is given it will list all snapshots of particular volume
func (d *CSIISCSiDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeDesc, maxret int, token *CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListVolumeSnapshots",
		"proto":   "iSCSI",
		"section": "driver",
	})

	entries, tnew, err := listEntries(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceSnapshotShort, jrest.RestError) {
			return d.re.GetVolumeSnapshotsEntries(ctx, pool, vid.VDS(), page, perPage)
		},
		func(s jrest.ResourceSnapshotShort) string { return s.Name },
		sameSnapshot,
		func(s jrest.ResourceSnapshotShort) bool { return d.ownSnapshot(s) && ownSnapshotName(l, vid, s.Name) },
		func(s jrest.ResourceSnapshotShort) string { return keyPart(s.Name) },
		token, maxret)
	if err != nil {
		return nil, nil, err
	}
	return entries, tnew, nil
}

// volumeSnapshots reads every snapshot of one volume, with no token and no
// limit, and whatever its name: the cleanup looks at the snapshots the
// plugin did not make too.
func (d *CSIISCSiDriver) volumeSnapshots(ctx context.Context, pool string, vd *VolumeDesc) ([]jrest.ResourceSnapshotShort, jrest.RestError) {
	read, err := readList(ctx,
		func(ctx context.Context, page, perPage int) ([]jrest.ResourceSnapshotShort, jrest.RestError) {
			return d.re.GetVolumeSnapshotsEntries(ctx, pool, vd.VDS(), page, perPage)
		})
	if err != nil {
		return nil, err
	}
	var snaps []jrest.ResourceSnapshotShort
	for _, s := range mergePages(read, func(s jrest.ResourceSnapshotShort) string { return s.Name }, sameSnapshot) {
		if d.ownSnapshot(s) {
			snaps = append(snaps, s)
		}
	}
	return snaps, nil
}

func NewJovianDSSCSIISCSiDriver(cfg *jcom.RestEndpointCfg, l *logrus.Entry) (d CSIDriver, err error) {
	var drvr CSIISCSiDriver
	jrest.SetupEndpoint(&drvr.re, cfg, l)

	return &drvr, nil
}

func (d *CSIISCSiDriver) GetVolume(ctx context.Context, pool string, vd *VolumeDesc) (out *GeneralVolume, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetVolume",
		"proto":   "iSCSI",
		"section": "driver",
	})

	l.Debugf("Get volume with id: %s", vd.VDS())

	if v, err := d.re.GetVolume(ctx, pool, vd.VDS()); err != nil {
		return nil, err
	} else {

		gv := GeneralVolume{
			Name:      v.Name,
			Size:      v.GetSize(),
			OVolume:   v.OriginVolume(),
			OSnapshot: v.OriginSnapshot(),
		}
		return &gv, nil
	}
}

func (d *CSIISCSiDriver) GetSnapshot(ctx context.Context, pool string, vd LunDesc, sd *SnapshotDesc) (out *GeneralSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetSnapshot",
		"proto":   "iSCSI",
		"section": "driver",
	})

	l.Debugf("Get snapshot %s of volume %s", sd.SDS(), vd.VDS())

	if nv, err := d.re.GetVolumeSnapshot(ctx, pool, vd.VDS(), sd.SDS()); err != nil {
		return nil, err
	} else {
		gv := GeneralSnapshot{
			// the reply of a zvol snapshot read by name carries no name:
			// the name is the one that was asked for
			Name:     sd.SDS(),
			VolSize:  nv.GetSize(),
			Creation: nv.GetCreation(),
		}
		return &gv, nil
	}
}

func (d *CSIISCSiDriver) CreateSnapshot(ctx context.Context, pool string, vd *VolumeDesc, sd *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "CreateSnapshot",
		"proto":   "iSCSI",
		"section": "driver",
	})

	l.Debugf("Create snapshot %s for volume %s", sd.SDS(), vd.VDS())

	snapdata := jrest.CreateSnapshotDescriptor{SnapshotName: sd.SDS()}

	return d.re.CreateSnapshot(ctx, pool, vd.VDS(), &snapdata)
}

func (d *CSIISCSiDriver) DeleteSnapshot(ctx context.Context, pool string, ld LunDesc, sd *SnapshotDesc) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "DeleteSnapshot",
		"proto":   "iSCSI",
		"section": "driver",
	})

	l.Debugf("Delete snapshot %s for volume %s", sd.SDS(), ld.VDS())

	// force_umount false: a delete never detaches a zvol from its target;

	// only DettachVolumeFromTarget does that.

	forceUmount := false
	deldata := jrest.DeleteSnapshotDescriptor{ForceUmount: &forceUmount}

	err := d.re.DeleteSnapshot(ctx, pool, ld.VDS(), sd.SDS(), deldata)

	var dvols []string
	var dsnaps []string
	var ncsi []string
	var msg string

	if err.GetCode() == jrest.RestErrorResourceBusy || err.GetCode() == jrest.RestErrorResourceBusySnapshotHasClones {
		if clones, rErr := d.re.GetVolumeSnapshotClones(ctx, pool, ld.VDS(), sd.SDS()); rErr != nil {
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
	} else {
		return err
	}

	if len(dsnaps) > 0 {
		// force_umount false: a delete never detaches a zvol from its target;
		// only DettachVolumeFromTarget does that.
		forceUmount := false
		delclone := jrest.DeleteVolumeDescriptor{ForceUmount: &forceUmount}

		for _, snapclone := range dsnaps {
			rErr := d.re.DeleteClone(ctx, pool, ld.VDS(), sd.SDS(), snapclone, delclone)
			switch jrest.ErrCode(rErr) {
			case jrest.RestErrorOk:
			case jrest.RestErrorResourceDNE:
				// the clone is gone already: nothing is in the way
			case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusyVolumeHasSnapshots:
				// the clone has clones or snapshots of its own
				msg = fmt.Sprintf("Unable to delete snapshot %s with ID %s because it has volume associated with it %s that cant be deleted, please delete physical zvol first", sd.Name(), sd.CSIID(), snapclone)
				return jrest.GetError(jrest.RestErrorResourceBusy, msg)
			default:
				// any other failure keeps its own meaning: a lost
				// connection is not "busy"
				return rErr
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

func (d *CSIISCSiDriver) GetPool(ctx context.Context, pool string) (out *jrest.ResourcePool, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetPool",
		"proto":   "iSCSI",
		"section": "driver",
	})

	l.Debugf("Get pool with id: %s", pool)

	return d.re.GetPool(ctx, pool)
}

func (d *CSIISCSiDriver) PublishVolume(ctx context.Context, pool string, ld LunDesc, iqnPrefix string, readonly bool) (iscsiContext *map[string]string, rErr jrest.RestError) {
	// Create target

	iContext := map[string]string{}

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "PublishVolume",
		"proto":   "iSCSI",
		"section": "driver",
	})

	// We want target name to be uniquee
	tname := fmt.Sprintf("%x", sha256.Sum256([]byte(ld.VDS())))
	iqn := fmt.Sprintf("%s:%s", iqnPrefix, tname)

	if len(tname) > 255 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Resulting target name is too long %s", tname))
	}

	var ctDesc jrest.CreateTargetDescriptor
	active := true
	ctDesc.Name = iqn
	ctDesc.Active = &active

	rErr = d.re.CreateTarget(ctx, pool, &ctDesc)

	switch jrest.ErrCode(rErr) {
	case jrest.RestErrorOk:
		l.Debugf("target %s created", tname)
	case jrest.RestErrorResourceExists:
		l.Debugf("target %s already exists", tname)
	default:
		return nil, rErr
	}

	// Attach to target
	var mode string = "wt"
	if readonly == true {
		mode = "ro"
	}

	var attachLun jrest.TargetLunDescriptor

	attachLun.Name = ld.VDS()
	attachLun.Mode = &mode
	lunID := 0
	attachLun.LUN = &lunID

	rErr = d.re.AttachVolumeToTarget(ctx, pool, iqn, &attachLun)

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case jrest.RestErrorResourceDNEVolume:
			d.re.DeleteTarget(ctx, pool, tname)
			return nil, rErr
		case jrest.RestErrorResourceBusy:
			// According to specification from
			// TODO: check that resource indeed properly assigned and continue if everything is ok
			l.Debugf("Volume %s already attached", ld.Name())
		default:
			return nil, rErr
		}
	}

	iContext["iqn"] = iqn
	iContext["target"] = tname
	iContext["lun"] = fmt.Sprintf("%d", lunID)

	for i := 0; i < 3; i++ {
		target, rErr := d.re.GetTarget(ctx, pool, iqn)
		switch jrest.ErrCode(rErr) {
		case jrest.RestErrorOk:
			if target.Active == true {
				return &iContext, nil
			}
		case jrest.RestErrorResourceDNE:
			// According to specification from
			time.Sleep(time.Second)
			continue
		default:
			continue
		}
	}

	return nil, jrest.GetError(jrest.RestErrorRequestTimeout, fmt.Sprintf("Unable to ensure that target %s is up and running", iqn))
}

func (d *CSIISCSiDriver) UnpublishVolume(ctx context.Context, pool string, prefix string, ld LunDesc) (rErr jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "UnpublishVolume",
		"section": "driver",
	})

	// We want target name to be uniquee
	iqn, rErr := TargetIQN(prefix, ld)

	if rErr != nil {
		return rErr
	}

	rErr = d.re.DettachVolumeFromTarget(ctx, pool, *iqn, ld.VDS())

	if rErr != nil {
		code := rErr.GetCode()
		switch code {
		case jrest.RestErrorOk:
			l.Debugf("Volume %s was detached from target %s", ld.Name(), *iqn)
		case jrest.RestErrorResourceDNELun, jrest.RestErrorResourceDNEVolume, jrest.RestErrorResourceDNE:
			// The volume is not attached, but the target can still be
			// there: go on and delete it.
			l.Debugf("Volume %s is not attached from target %s", ld.Name(), *iqn)
		case jrest.RestErrorResourceDNETarget:
			// The target is gone: nothing is left to do.
			return nil
		default:
			return rErr
		}
	}

	rErr = d.re.DeleteTarget(ctx, pool, *iqn)

	switch jrest.ErrCode(rErr) {
	case jrest.RestErrorOk:
		l.Debugf("target %s deleted", *iqn)
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNETarget:
		l.Debugf("target %s do not exists", *iqn)
	default:
		return rErr
	}

	return nil
}
