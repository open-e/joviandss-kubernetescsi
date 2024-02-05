package driver

import (

	"fmt"

	"golang.org/x/net/context"
	"github.com/sirupsen/logrus"

	jrest "joviandss-kubernetescsi/pkg/rest"
	jcom "joviandss-kubernetescsi/pkg/common"

)

// JovianDSS CSI plugin
type CSIDriver struct {
	re		jrest.RestEndpoint
	l		*logrus.Entry
}


func (d *CSIDriver) cloneLUN(ctx context.Context, source LunID, dest LunID) error {


	return nil
}

func (d *CSIDriver) CreateVolume(ctx context.Context, pool string, nvid *VolumeId, volumeSize int64) error {
	
	vd := jrest.CreateVolumeDescriptor{
		Name: nvid.VID(),
		Size: fmt.Sprintf("%d", volumeSize),
	}

	err := d.re.CreateVolume(ctx, pool, vd)

	return err
}

func (d *CSIDriver) CreateVolumeFromSnapshot(ctx context.Context, sid *SnapshotId, nvid *VolumeId) error {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "createVolumeFromSnapshot",
	})

	d.cloneLUN(ctx, sid, nvid)
//
//	snameT := strings.Split(sname, "_")
//	var vname string
//	if len(snameT) == 2 {
//		vname = snameT[0]
//	} else if len(snameT) == 3 {
//		vname = snameT[1]
//	} else {
//		msg := "Unable to obtain volume name from snapshot name"
//		l.Warn(msg)
//		return status.Error(codes.NotFound, msg)
//	}
//
//	rErr := (*cp.endpoints[0]).CreateClone(vname, sname, nvname)
//	var err error
//	if rErr != nil {
//		switch rErr.GetCode() {
//		case rest.RestRequestMalfunction:
//			// TODO: correctly process error messages
//			err = status.Error(codes.NotFound, rErr.Error())
//			// return nil, status.Error(codes.Internal, rErr.Error())
//		case rest.RestObjectExists:
//			err = status.Error(codes.FailedPrecondition, rErr.Error())
//		case rest.RestRPM:
//			err = status.Error(codes.Internal, rErr.Error())
//		case rest.RestResourceDNE:
//			err = status.Error(codes.NotFound, rErr.Error())
//		default:
//			err = status.Errorf(codes.Internal, rErr.Error())
//		}
//		return err
//	}
//
	return nil
}

func (d *CSIDriver) CreateVolumeFromVolume(ctx context.Context, vid *VolumeId, nvid *VolumeId) error {

	l := jcom.LFC(ctx)
	l = l.WithFields(logrus.Fields{
		"func": "CreateVolumeFromVolume",
	})

	d.cloneLUN(ctx, vid, nvid)
	msg := fmt.Sprintf("Create %s From %s", vid.ID(), nvid.ID())
	l.Tracef(msg)

	err := d.cloneLUN(ctx, vid, nvid)
	return err
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

