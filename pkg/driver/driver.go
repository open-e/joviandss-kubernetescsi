package driver

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	jcom "joviandss-kubernetescsi/pkg/common"
	jrest "joviandss-kubernetescsi/pkg/rest"
)

// JovianDSS CSI plugin
type CSIDriver struct {
	re		jrest.RestEndpoint
	l		*logrus.Entry
}


func (d *CSIDriver) cloneLUN(ctx context.Context, pool string, source LunDesc, clone LunDesc, snap *SnapshotDesc) jrest.RestError {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "cloneLUN",
	})
	
	l.Debugf("Start cloning")

	var sds string
	if snap == nil {
		var snapdata = jrest.CreateSnapshotDescriptor{SnapshotName: clone.VDS()} 

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
	}else {
		sds = snap.sds
	}

	var clonedata = jrest.CloneVolumeDescriptor{Name: clone.VDS(), Snapshot: sds }
	err := d.re.CreateClone(ctx, pool, source.VDS(), clonedata)
	return err
}

func (d *CSIDriver) CreateVolume(ctx context.Context, pool string, nvd *VolumeDesc, volumeSize int64) jrest.RestError {

	vd := jrest.CreateVolumeDescriptor{
		Name: nvd.VDS(),
		Size: fmt.Sprintf("%d", volumeSize),
	}

	err := d.re.CreateVolume(ctx, pool, vd)

	return err
}

func (d *CSIDriver) CreateVolumeFromSnapshot(ctx context.Context, pool string, sd *SnapshotDesc, nvd *VolumeDesc) jrest.RestError {
	
	var clonedata = jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: sd.SDS() }
	return d.re.CreateClone(ctx, pool, sd.ld.VDS(), clonedata)
}


func (d *CSIDriver) deleteIntermediateSnapshot(ctx context.Context, pool string, vds string, sds string) (err jrest.RestError) {


	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "deleteIntermediateSnapshot",
		"section": "driver",
	})
	snapdeldata := jrest.DeleteSnapshotDescriptor{ForceUnmount:true}
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
					return  jrest.GetError(jrest.RestErrorStorageFailureUnknown,
								fmt.Sprintf("It looks like there is a race condition"))
				}
				return jrest.GetError(jrest.RestErrorStorageFailureUnknown,
							fmt.Sprintf("Unable to delete intermediate snapshot %s for volume as clean up operation because of %s",
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
func (d *CSIDriver) CreateVolumeFromVolume(ctx context.Context, pool string, vd *VolumeDesc, nvd *VolumeDesc) (err jrest.RestError) {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "CreateVolumeFromVolume",
		"section": "driver",
	})

	var snapdata = jrest.CreateSnapshotDescriptor{SnapshotName: nvd.VDS()}

	if err := d.re.CreateSnapshot(ctx, pool, vd.VDS(), &snapdata); err != nil {
		code := err.GetCode()
		// We are not able to create this snapshot for some reason

		// Probably it was already created
		if code == jrest.RestErrorResourceExists {
		
			d.deleteIntermediateSnapshot(ctx, pool, vd.VDS(), nvd.VDS()) 
		}
		return err
	}

	var clonedata = jrest.CloneVolumeDescriptor{Name: nvd.VDS(), Snapshot: nvd.VDS() }
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
func (d *CSIDriver) cleanIntermediateSnapshots(ctx context.Context, pool string, vd *VolumeDesc) (snaps []jrest.ResourceSnapshot, err jrest.RestError) {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "cleanIntermediateSnapshots",
		"section": "driver",
	})

	snaps, _, gserr := d.ListVolumeSnapshots(ctx, pool, vd, nil, nil)

	l.Debugf("Clean intermediate snapshots return %d records", len(snaps))
	if gserr != nil {
		l.Debugf("Unable to get list of snapshots for volume %s", vd.Name())
		return nil, gserr
	}

	var out []jrest.ResourceSnapshot

	for _, snap := range snaps {
		if IsVDS(snap.Name) {
			clones := snap.ClonesNames()
			if len(clones) == 0 {
				snapdeldata := jrest.DeleteSnapshotDescriptor{ForceUnmount:true}

				err = d.re.DeleteSnapshot(ctx, pool, vd.VDS(), snap.Name, snapdeldata)
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


func (d *CSIDriver) deleteLUN(ctx context.Context, pool string, vd *VolumeDesc) (err jrest.RestError) {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "deleteLUN",
		"section": "driver",
	})

	var deldata = jrest.DeleteVolumeDescriptor{ ForceUmount: true, RecursivelyChildren: true }
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
		var ncsi[]string
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
			msg += fmt.Sprintf(" not CSI relates snapshots: %s", strings.Join(ncsi[:], ","))
		}
		err = jrest.GetError(jrest.RestErrorResourceBusy, msg)
			//fmt.Sprintf("%s %v %v", msg, dsnaps, dvols))
	}

	return err
}

func (d *CSIDriver) DeleteVolume(ctx context.Context, pool string, vid *VolumeDesc) jrest.RestError {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "DeleteVolume",
		"section": "driver",
	})

	return d.deleteLUN(ctx, pool, vid)
}

func (d *CSIDriver) findPageByToken(ctx context.Context, pool string, token *string) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "findPageByToken",
	})

	l.Debugf("Start volume deletion")
	
	return nil
}


// func (d *CSIDriver) getSnapshotsSlice(ctx context.Context, pool string, st *string, n int64) (t *string, snaps *[]jrest.ResourceSnapshot, err jrest.RestError) {
// 
// 	if st != nil {
// 		curentToken = NewSnapshotToken(st)
// 	}
// 	err := d.re.GetVolumeSnapshots(ctx, pool, vid.VID(), deldata)
// 
// 	return nil, nil, nil
// }

// func (d *CSIDriver) processSlice[T any](s []T) {
//     for _, item := range s {
//         fmt.Println(item)
//     }
// }


func getResourcesList[RestResource any](ctx context.Context, maxret uint64, token CSIListingToken, 
	grf func(ctx context.Context, token CSIListingToken)(lres []RestResource, err jrest.RestError),
	BasedID func(res RestResource) string) (lres []RestResource, nt *CSIListingToken, err jrest.RestError) {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "getResourceList",
		"section": "driver",
	})

	l.Debugf("Processing %T", lres)

	for {
		if ent, err := grf(ctx, token); err != nil {
			return nil, nil, err
		} else {
			// No new snapshots, return what we have so far
			//data, ok := ent.Entries.()

			if len(ent) == 0 {
				return lres, nil, nil
			}

			if token.BasedID() <= BasedID(ent[0]) {
				lres = append(lres, (ent)...)
				token.DropBasedID()
			} else {
				for i, e := range ent {
					if BasedID(e) ==  token.BasedID() {
						lres = append(lres, ent[i:]...)
						token.BasedID()
						break
					}
				}
			}
		}

		if maxret > 0 {
			if uint64(len(lres)) >= maxret {
				if newToken, err := NewCSIListingTokenFromBasedID(BasedID(lres[maxret]), token.Page(), token.DC()); err != nil {
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

func (d *CSIDriver) ListAllSnapshots(ctx context.Context, pool string, maxret uint64, token CSIListingToken) (snaps []jrest.ResourceSnapshotShort, tnew *CSIListingToken, err jrest.RestError) {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "ListAllSnapshots",
		"section": "driver",
	})

	grf := func(ctx context.Context, token CSIListingToken)(lres []jrest.ResourceSnapshotShort, err jrest.RestError) {
	 	entr, err := d.re.GetSnapshotsEntries(ctx, pool, token.Page(), token.DC())
		
		if err != nil {
			return nil, err
		}
		
		if entries, ok := entr.Entries.([]jrest.ResourceSnapshotShort); ok == true {
			return entries, nil
		}
		l.Warnln("Unable to identify format of %+v, it have %T")
		return nil, nil
	}

	//var token CSIListingToken
	//if startToken != nil {
	//	if t, err := NewCSIListingTokenFromTokenString(*startToken); err != nil {
	//		return nil, nil, err
	//	} else {
	//		token = *t
	//	}
	//}else {
	//	token = NewCSIListingToken()
	//}

	if entries, csitoken, err := getResourcesList(ctx, maxret, token,  grf, RestSnapshotShortEntryBasedID); err != nil {
		return nil, nil, err
	} else {
		//ts := csitoken.Token()
		return entries, csitoken , nil
	}

	//rsnaps := []jrest.ResourceSnapshot{}
	//var token *snapshotToken
	//if startToken != nil {
	//	if token, err = NewSnapshotTokenFromStr(*startToken); err != nil {
	//		return nil, nil, err
	//	}
	//} else {
	//	token = NewSnapshotToken(0, rand.Int63(), "", "")
	//}
	//for {
	//	if _, spage, err := d.re.GetVolumeSnapshots(ctx, pool, vid.VID(), &token.page, &token.dc); err != nil {
	//		return nil, nil, err
	//	} else {
	//		// No new snapshots, return what we have so far
	//		if len(*spage) == 0 {
	//			return &rsnaps, nil, nil
	//		}

	//		if token.sid <= (*spage)[0].Name {
	//			rsnaps = append(rsnaps, (*spage)...)
	//			token.sid = ""
	//			token.vid = ""
	//		} else {
	//			for i, snap := range *spage {
	//				if snap.Name == token.sid  {
	//					rsnaps = append(rsnaps, (*spage)[i:]...)
	//					token.sid = ""
	//					token.vid = ""
	//					break
	//				}
	//			}
	//		}
	//	}

	//	if maxret != nil {
	//		if int64(len(*snaps)) >= *maxret {
	//			newToken := NewSnapshotToken(token.page, token.dc, "", rsnaps[*maxret].Name)
	//			rsnaps = rsnaps[:*maxret]
	//			nts := newToken.String()
	//			return &rsnaps, &nts, nil
	//		}
	//	}
	//	token.page += 1
	//}

	//return nil
}


// func (d *CSIDriver) ListAllSnapshots(ctx context.Context, pool string, maxret *int64, startToken *string) (snaps *[]jrest.ResourceSnapshot, tnew *string, err jrest.RestError) {
// 	
// 	rsnaps := []jrest.ResourceSnapshot{}
// 	var token *snapshotToken
// 	if startToken != nil {
// 		if token, err = NewSnapshotTokenFromStr(*startToken); err != nil {
// 			return nil, nil, err
// 		}
// 	} else {
// 		token = NewSnapshotToken(0, rand.Int63(), "", "")
// 	}
// 	for {
// 		if _, spage, err := d.re.GetVolumeSnapshots(ctx, pool, vid.VID(), &token.page, &token.dc); err != nil {
// 			return nil, nil, err
// 		} else {
// 			// No new snapshots, return what we have so far
// 			if len(*spage) == 0 {
// 				return &rsnaps, nil, nil
// 			}
// 
// 			if token.sid <= (*spage)[0].Name {
// 				rsnaps = append(rsnaps, (*spage)...)
// 				token.sid = ""
// 				token.vid = ""
// 			} else {
// 				for i, snap := range *spage {
// 					if snap.Name == token.sid  {
// 						rsnaps = append(rsnaps, (*spage)[i:]...)
// 						token.sid = ""
// 						token.vid = ""
// 						break
// 					}
// 				}
// 			}
// 		}
// 
// 		if maxret != nil {
// 			if int64(len(*snaps)) >= *maxret {
// 				newToken := NewSnapshotToken(token.page, token.dc, "", rsnaps[*maxret].Name)
// 				rsnaps = rsnaps[:*maxret]
// 				nts := newToken.String()
// 				return &rsnaps, &nts, nil
// 			}
// 		}
// 		token.page += 1
// 	}
// 
// 	return nil
// }

// ListVolumeSnapshots provides maxret records of snapshots of volume starting from token
// if no token nor limit on number of snapshot is given it will list all snapshots of particular volume
func (d *CSIDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeDesc, maxret *int64, tcur *string) (snaps []jrest.ResourceSnapshot, tnew *string, err jrest.RestError) {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "ListVolumeSnapshots",
		"section": "driver",
	})

	rsnaps := []jrest.ResourceSnapshot{}
	// Init toke that will be used to iterate over response from storage
	var token *snapshotToken
	if tcur != nil {
		if token, err = NewSnapshotTokenFromStr(*tcur); err != nil {
			return nil, nil, err
		}
	} else {
		// No token is given so we start from page 0
		token = NewSnapshotToken(0, rand.Int63(), "", "")
	}

	for {
		if _, spage, err := d.re.GetVolumeSnapshots(ctx, pool, vid.VDS(), &token.page, &token.dc); err == nil {

			// No new snapshots, return what we have so far
			if len(*spage) == 0 {
				return rsnaps, nil, nil
			}

			if token.sid <= (*spage)[0].Name {
				rsnaps = append(rsnaps, (*spage)...)
				token.sid = ""
				token.vid = ""
			} else {
				for i, snap := range *spage {
					if snap.Name == token.sid  {
						rsnaps = append(rsnaps, (*spage)[i:]...)
						token.sid = ""
						token.vid = ""
						break
					}
				}
			}
		} else {
			// Unable to get volume snapshots
			l.Debugf("Unable to get snapshots for volume %s because of %s", vid.VDS(), err.Error())
			return nil, nil, err
		}

		if maxret != nil {
			if int64(len(snaps)) >= *maxret {
				if int64(len(snaps)) == *maxret {
					token.page += 1
				}
				newToken := NewSnapshotToken(token.page, token.dc, "", rsnaps[*maxret].Name)
				rsnaps = rsnaps[:*maxret]
				nts := newToken.String()
				return rsnaps, &nts, nil
			} 
		}
		token.page += 1
	}
}

func NewJovianDSSCSIDriver(cfg *jcom.RestEndpointCfg, l *logrus.Entry) (d *CSIDriver, err error) {
	
	var drvr CSIDriver
	jrest.SetupEndpoint(&drvr.re, cfg, l)

	return &drvr, nil 
}

func (d* CSIDriver)GetVolume(ctx context.Context, pool string, vd *VolumeDesc) (out *jrest.ResourceVolume, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "GetVolume",
		"section" : "driver",
	})

	l.Debugf("Get volume with id: %s", vd.VDS())
	

	return d.re.GetVolume(ctx, pool, vd.VDS()) // v for Volume

	// if rErr != nil {
	// 	switch rErr.GetCode() {
	// 	case rest.RestRequestMalfunction:
	// 		// TODO: correctly process error messages
	// 		err = status.Error(codes.NotFound, rErr.Error())

	// 	case rest.RestRPM:
	// 		err = status.Error(codes.Internal, rErr.Error())
	// 	case rest.RestResourceDNE:
	// 		err = status.Error(codes.NotFound, rErr.Error())
	// 	default:
	// 		err = status.Errorf(codes.Internal, rErr.Error())
	// 	}
	// 	return nil, err
	// }
	//return v, nil
}

func (d* CSIDriver)GetSnapshot(ctx context.Context, pool string, vd *VolumeDesc, sd *SnapshotDesc) (out *jrest.ResourceSnapshot, err jrest.RestError) {
	// return nil, nil

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "GetSnapshot",
		"section" : "driver",
	})

	l.Debugf("Get snapshot %s of volume %s", sd.SDS(), vd.VDS())
	

	return d.re.GetVolumeSnapshot(ctx, pool, vd.VDS(), sd.SDS())
}

