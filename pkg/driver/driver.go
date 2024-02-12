package driver

import (
	"fmt"
	"math/rand"

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


func (d *CSIDriver) cloneLUN(ctx context.Context, pool string, source LunID, dest LunID) jrest.RestError {
	
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "cloneLUN",
	})
	
	l.Debugf("Start cloning")

	var snapdata = jrest.CreateSnapshotDescriptor{SnapshotName: dest.VID()} 

	if err := d.re.CreateSnapshot(ctx, pool, source.VID(), &snapdata); err != nil {
		code := err.GetCode()
		if code != jrest.RestErrorResourceExists {
			return err
		}
		// TODO: check if specific snapshot have clones, if it does and name of clone
		//	is the same as name of the clone, then return error
		//	if there are no clones snapshot have to be deleteted and recreated
	}

	var clonedata = jrest.CloneVolumeDescriptor{Name: dest.VID(), Snapshot:dest.VID() }
	err := d.re.CreateClone(ctx, pool, source.VID(), clonedata)
	return err
}

func (d *CSIDriver) CreateVolume(ctx context.Context, pool string, nvid *VolumeId, volumeSize int64) jrest.RestError {
	
	vd := jrest.CreateVolumeDescriptor{
		Name: nvid.VID(),
		Size: fmt.Sprintf("%d", volumeSize),
	}

	err := d.re.CreateVolume(ctx, pool, vd)

	return err
}

func (d *CSIDriver) CreateVolumeFromSnapshot(ctx context.Context, pool string, sid *SnapshotId, nvid *VolumeId) jrest.RestError {

	return d.cloneLUN(ctx, pool, sid, nvid)
}

func (d *CSIDriver) CreateVolumeFromVolume(ctx context.Context, pool string, vid *VolumeId, nvid *VolumeId) jrest.RestError {

	return d.cloneLUN(ctx, pool, vid, nvid)
}


func (d *CSIDriver) deleteLUN(ctx context.Context, pool string, vid LunID) jrest.RestError {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "deleteLUN",
	})

	l.Debugf("Start volume deletion")

	var deldata = jrest.DeleteVolumeDescriptor{ ForceUmount: true, RecursivelyChildren: true }
	err := d.re.DeleteVolume(ctx, pool, vid.VID(), deldata)

	switch err.GetCode() {
	case jrest.RestErrorResourceBusy:
		break
	case jrest.RestErrorResourceDNE:
		return nil
	default:
		return err
	}

	//d.re.get
	return nil
}

func (d *CSIDriver) DeleteVolume(ctx context.Context, pool string, vid *VolumeId) jrest.RestError {

	d.deleteLUN(ctx, pool, vid)
	return nil
}

func (d *CSIDriver) findPageByToken(ctx context.Context, pool string, token *string) jrest.RestError {
	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "findPageByToken",
	})

	l.Debugf("Start volume deletion")
	
	return nil
}


func (d *CSIDriver) getSnapshotsSlice(ctx context.Context, pool string, st *string, n int64) (t *string, snaps *[]jrest.ResourceSnapshot, err jrest.RestError) {

	if st != nil {
		curentToken = NewSnapshotToken(st)
	}
	err := d.re.GetVolumeSnapshots(ctx, pool, vid.VID(), deldata)

	return nil, nil, nil
}

func (d *CSIDriver) ListSnapshots(ctx context.Context, pool string, maxret *int64, token *string) jrest.RestError {


	return nil
}

func (d *CSIDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *VolumeId, maxret *int64, tcur *string) (snaps *[]jrest.ResourceSnapshot, tnew *string, err jrest.RestError) {

	rsnaps := []jrest.ResourceSnapshot{}
	var token *snapshotToken
	if tcur != nil {
		if token, err = NewSnapshotTokenFromStr(*tcur); err != nil {
			return nil, nil, err
		}
	} else {
		token = NewSnapshotToken(0, rand.Int63(), "", "")
	}
	for {
		if nsnaps, spage, err := d.re.GetVolumeSnapshots(ctx, pool, vid.VID(), &token.page, &token.dc); err != nil {
			return nil, nil, err
		} else {
			if &token.sid0
			rsnaps = append(rsnaps, *spage...)
		}

		if maxret != nil {
			if int64(len(*snaps)) >= *maxret {
				newToken := NewSnapshotToken(token.page, token.dc, "", rsnaps[*maxret].Name)
				rsnaps = rsnaps[:*maxret]
				nts := newToken.String()
				return &rsnaps, &nts, nil
			}
		}
	}
	return nil, nil, nil
}

// func SetupJovianDSSDriver( l *logrus.Logger) (
// 	err error,
// ) {
// 	return 
// }

// func GetVolume(vID string) (*jrest.Volume, error) {
// 	// return nil, nil
// 	l := cp.l.WithFields(logrus.Fields{
// 		"func": "getVolume",
// 	})
// 
// 	l.Tracef("Get volume with id: %s", vID)
// 	var err error
// 
// 	//////////////////////////////////////////////////////////////////////////////
// 	/// Checks
// 
// 	if len(vID) == 0 {
// 		msg := "Volume name missing in request"
// 		l.Warn(msg)
// 		return nil, status.Error(codes.InvalidArgument, msg)
// 	}
// 
// 	//////////////////////////////////////////////////////////////////////////////
// 
// 	v, rErr := (*cp.endpoints[0]).GetVolume(vID) // v for Volume
// 
// 	if rErr != nil {
// 		switch rErr.GetCode() {
// 		case rest.RestRequestMalfunction:
// 			// TODO: correctly process error messages
// 			err = status.Error(codes.NotFound, rErr.Error())
// 
// 		case rest.RestRPM:
// 			err = status.Error(codes.Internal, rErr.Error())
// 		case rest.RestResourceDNE:
// 			err = status.Error(codes.NotFound, rErr.Error())
// 		default:
// 			err = status.Errorf(codes.Internal, rErr.Error())
// 		}
// 		return nil, err
// 	}
// 	return v, nil
// }

