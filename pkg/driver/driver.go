package driver

import (
	jrest "joviandss-kubernetescsi/pkg/rest"
	"github.com/sirupsen/logrus"
)

// JovianDSS CSI plugin
type JovianDSSDriver struct {
	j    *jrest.Storage
	l    *logrus.Entry
}

func SetupJovianDSSDriver( l *logrus.Logger) (
	err error,
) {

}

func GetVolume(vID string) (*jrest.Volume, error) {
	// return nil, nil
	l := cp.l.WithFields(logrus.Fields{
		"func": "getVolume",
	})

	l.Tracef("Get volume with id: %s", vID)
	var err error

	//////////////////////////////////////////////////////////////////////////////
	/// Checks

	if len(vID) == 0 {
		msg := "Volume name missing in request"
		l.Warn(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	//////////////////////////////////////////////////////////////////////////////

	v, rErr := (*cp.endpoints[0]).GetVolume(vID) // v for Volume

	if rErr != nil {
		switch rErr.GetCode() {
		case rest.RestRequestMalfunction:
			// TODO: correctly process error messages
			err = status.Error(codes.NotFound, rErr.Error())

		case rest.RestRPM:
			err = status.Error(codes.Internal, rErr.Error())
		case rest.RestResourceDNE:
			err = status.Error(codes.NotFound, rErr.Error())
		default:
			err = status.Errorf(codes.Internal, rErr.Error())
		}
		return nil, err
	}
	return v, nil
}

