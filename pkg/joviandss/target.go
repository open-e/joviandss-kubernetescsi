package joviandss

import (
	"github.com/sirupsen/logrus"
)

// Target stores info about iscsi target
type Target struct {
	l          *logrus.Entry
	cfg        *NodeCfg
	STPath     string // Where target is staged
	TPath      string // Where target should be mounted
	DPath      string // Device representation in system
	Portal     string // ip of JovianDSS
	PortalPort string // port of JovianDSS
	Iqn        string // prefix part of iqn
	Lun        string // expected to be 0
	Tname      string // target name = volumeID
	CoUser     string // Chap outgoing password
	CoPass     string // Chap outgoing Password
	TProtocol  string // tcp, others are not supported

	FsType     string   // Type of file system
	MountFlags []string // mount tool arguments
}
