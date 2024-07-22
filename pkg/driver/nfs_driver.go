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

	enable_NFS := true
	all_squash_NFS := true
	active_share := true

	path := fmt.Sprintf("%s/%s/%s", pool, name, name)
	l.Debugf("Create share with path %s", path)

	nfs := jrest.ShareNFSDescriptor{
		Enabled:   &enable_NFS,
		AllSquash: &all_squash_NFS,
	}

	sd := jrest.CreateShareDescriptor{
		Name:   name,
		Active: &active_share,
		Path:   path,
		NFS:    &nfs,
	}

	return d.re.CreateShare(ctx, &sd)
}

func (d *CSINFSDriver) CreateVolumeFromSnapshot(ctx context.Context, pool string, sd *SnapshotDesc, nvd *VolumeDesc) jrest.RestError {
	// clonedata := jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: sd.SDS()}
	// TODO: implement create from clone

	return jrest.GetError(jrest.RestErrorFailureUnknown,
		fmt.Sprintf("Creation of volume from snapshot is not implemented"))
	// return &jrest.restError{code: jrest.RestErrorFailureUnknown}
	// return d.re.CreateClone(ctx, pool, sd.ld.VDS(), clonedata)
}

func (d *CSINFSDriver) deleteIntermediateSnapshot(ctx context.Context, pool string, vds string, sds string) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteIntermediateSnapshot",
		"section": "driver",
	})
	forceUmount := true
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

	// clonedata := jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: nvd.VDS()}
	// if err = d.re.CreateClone(ctx, pool, vd.VDS(), clonedata); err != nil {
	// 	l.Warnf("Unable to create volume %s from snapshot %s of volume %s, because of error %+v. Removing intermediate snapshot", nvd.VDS(), nvd.VDS(), vd.VDS(), err.Error())

	// 	d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS())
	// 	return err
	// }

	return nil
}

// cleanIntermediateSnapshots request list of snapshots related to particular volume and check if there are intermediate one that can be deleted
//
//	return list of snapshots that does not contain 'handing' one or error
func (d *CSINFSDriver) cleanIntermediateSnapshots(ctx context.Context, pool string, vd *VolumeDesc) (out []jrest.ResourceSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "cleanIntermediateSnapshots",
		"section": "driver",
	})
	token := NewCSIListingToken()
	isnaps, _, gserr := d.ListVolumeSnapshots(ctx, pool, vd, 0, token)

	var snaps []jrest.ResourceSnapshot

	if val, ok := isnaps.([]jrest.ResourceSnapshot); ok == true {
		snaps = val
	} else {
		err = jrest.GetError(jrest.RestErrorFailureUnknown, "Unable to typecast snapshot data")
		return nil, err
	}
	l.Debugf("Clean intermediate snapshots return %d records", len(snaps))
	if gserr != nil {
		l.Debugf("Unable to get list of snapshots for volume %s", vd.Name())
		return nil, gserr
	}

	for _, snap := range snaps {
		if IsVDS(snap.Name) {
			clones := snap.ClonesNames()
			if len(clones) == 0 {
				// TODO: think about arguments for deletion command
				// forceUmount := true
				// snapdeldata := jrest.DeleteSnapshotDescriptor{ForceUmount: &forceUmount}

				err = d.re.DeleteNASVolumeSnapshot(ctx, pool, vd.VDS(), snap.Name)
				if err != nil {
					return nil, err
				}
			} else {
				out = append(out, snap)
			}
		} else if IsSDS(snap.Name) {
			out = append(out, snap)
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

	err = d.re.DeleteShare(ctx, vd.VDS(), nil)
	switch jrest.ErrCode(err) {
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEShare:
		break
	case jrest.RestErrorOk:
		break
	default:
		l.Debugf("Unable to delete share %s, error had happaned %+v", vd.Name(), err.Error())
		return err
	}

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

func (d *CSINFSDriver) deleteLUN(ctx context.Context, pool string, vd *VolumeDesc) (err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "deleteLUN",
		"section": "driver",
	})

	err = d.deleteNASVolAndShare(ctx, pool, vd)

	switch jrest.ErrCode(err) {
	case jrest.RestErrorResourceBusy, jrest.RestErrorResourceBusyVolumeHasSnapshots:
		break
	case jrest.RestErrorResourceDNE, jrest.RestErrorResourceDNEVolume:
		l.Debugf("NAS volume %s do not exists. Do nothing.", vd.Name())
		return nil
	case jrest.RestErrorOk:
		return nil
	default:
		l.Debugf("Unable to delete lun %s, error had happaned %+v", vd.Name(), err.Error())
		return err
	}
	l.Debugf("Volume %s is busy, indentifying relaying resources", vd.Name())

	if err.GetCode() == jrest.RestErrorResourceBusy {
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
		// Looks like this volume is busy and we are not able to delete it

		var dvols []string
		var dsnaps []string
		var ncsi []string
		msg := fmt.Sprintf("Volume %s is dependent upon by", vd.Name())

		for _, snap := range snaps {
			if IsSDS(snap.Name) {
				dsnaps = append(dsnaps, snap.Name)
			} else if IsVDS(snap.Name) {
				clones := snap.ClonesNames()
				dvols = append(dvols, clones...)
			} else {
				msg += fmt.Sprintf(" not CSI relates snapshots: %s", strings.Join(dsnaps[:], ","))
			}
		}

		if len(dvols) > 0 {
			msg += fmt.Sprintf(" volumes: %s", strings.Join(dvols[:], ","))
		}
		if len(dsnaps) > 0 {
			msg += fmt.Sprintf(" snapshots: %s", strings.Join(dsnaps[:], ","))
		}
		if len(ncsi) > 0 {
			msg += fmt.Sprintf(" not CSI related snapshots: %s", strings.Join(ncsi[:], ","))
		}
		err = jrest.GetError(jrest.RestErrorResourceBusy, msg)
	}

	return err
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

func (d *CSINFSDriver) ListAllVolumes(ctx context.Context, pool string, maxret int, token CSIListingToken) (vols interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllVolumes",
		"section": "driver",
	})

	grf := func(ctx context.Context, token CSIListingToken) (lres []jrest.ResourceNASVolume, err jrest.RestError) {
		l.Debugln("Getting Volume entries")
		entr, err := d.re.GetNASVolumesEntries(ctx, pool, token.Page(), token.DC())
		if err != nil {
			return nil, err
		}

		if entries, ok := entr.Entries.(*[]jrest.ResourceNASVolume); ok == true {
			return *entries, nil
		}
		l.Warnf("Unable to identify format of %+v, it have %T", entr.Entries, entr.Entries)
		return nil, nil
	}

	if entries, csitoken, err := getResourcesList(ctx, maxret, token, grf, RestNASVolumeEntryBasedID); err != nil {
		return nil, nil, err
	} else {
		return entries, csitoken, nil
	}
}

func (d *CSINFSDriver) ListAllSnapshots(ctx context.Context, pool string, maxret int, token CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListAllSnapshots",
		"section": "driver",
	})

	grf := func(ctx context.Context, token CSIListingToken) (lres []jrest.ResourceNASSnapshotShort, err jrest.RestError) {
		l.Debugln("Getting SnapshotShort entries")
		entr, err := d.re.GetNASSnapshotsEntries(ctx, pool, token.Page(), token.DC())
		if err != nil {
			return nil, err
		}
		l.Debugf("Rest returned entries %+v", entr)

		if entries, ok := entr.Entries.(*[]jrest.ResourceNASSnapshotShort); ok == true {
			return *entries, nil
		}
		l.Warnf("Unable to identify format of %+v, it have %T", entr.Entries, entr.Entries)
		return nil, nil
	}

	if entries, csitoken, err := getResourcesList(ctx, maxret, token, grf, RestNASSnapshotShortEntryBasedID); err != nil {
		l.Warnf("Getting snapshots error %+v", err)
		return nil, nil, err
	} else {
		// ts := csitoken.Token()
		l.Debugf("Entries %+v", entries)
		return entries, csitoken, nil
	}
}

// ListVolumeSnapshots provides maxret records of snapshots of volume starting from token
// if no token nor limit on number of snapshot is given it will list all snapshots of particular volume
func (d *CSINFSDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeDesc, maxret int, token CSIListingToken) (snaps interface{}, tnew *CSIListingToken, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "ListNASVolumeSnapshots",
		"proto":   "NFS",
		"section": "driver",
	})

	grf := func(ctx context.Context, token CSIListingToken) (lres []jrest.ResourceNASVolumeSnapshot, err jrest.RestError) {
		l.Debugln("Getting Volume Snapshots entries")
		entr, err := d.re.GetNASVolumeSnapshotsEntries(ctx, pool, vid.VDS(), token.Page(), token.DC())
		if err != nil {
			return nil, err
		}

		if entries, ok := entr.Entries.(*[]jrest.ResourceNASVolumeSnapshot); ok == true {
			return *entries, nil
		}
		l.Warnf("Unable to identify format of %+v, it have %T", entr.Entries, entr.Entries)
		return nil, nil
	}

	if entries, csitoken, err := getResourcesList(ctx, maxret, token, grf, RestNASVolumeSnapshotEntryBasedID); err != nil {
		return nil, nil, err
	} else {
		return entries, csitoken, nil
	}
}

func NewJovianDSSCSINFSDriver(cfg *jcom.RestEndpointCfg, l *logrus.Entry) (d CSIDriver, err error) {
	var drvr CSINFSDriver
	jrest.SetupEndpoint(&drvr.re, cfg, l)

	return &drvr, nil
}

func (d *CSINFSDriver) GetVolume(ctx context.Context, pool string, vd *VolumeDesc) (out *jrest.ResourceVolume, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetVolume",
		"section": "driver",
	})

	l.Debugf("Get volume with id: %s", vd.VDS())

	return d.re.GetVolume(ctx, pool, vd.VDS()) // v for Volume
}

func (d *CSINFSDriver) GetSnapshot(ctx context.Context, pool string, vd LunDesc, sd *SnapshotDesc) (out *jrest.ResourceSnapshot, err jrest.RestError) {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func":    "GetSnapshot",
		"section": "driver",
	})

	l.Debugf("Get snapshot %s of volume %s", sd.SDS(), vd.VDS())

	return d.re.GetVolumeSnapshot(ctx, pool, vd.VDS(), sd.SDS())
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

	if err.GetCode() == jrest.RestErrorResourceBusy || err.GetCode() == jrest.RestErrorResourceBusySnapshotHasClones {
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
	} else {
		return err
	}

	if len(dsnaps) > 0 {
		forceUmount := true
		delclone := jrest.DeleteVolumeDescriptor{ForceUmount: &forceUmount}

		for _, snapclone := range dsnaps {
			if rErr := d.re.DeleteClone(ctx, pool, ld.VDS(), sd.SDS(), snapclone, delclone); rErr != nil {
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
