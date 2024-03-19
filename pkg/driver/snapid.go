package driver

import (
	"fmt"
	"strings"
	"encoding/base64"
	"crypto/sha256"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "joviandss-kubernetescsi/pkg/common"
	//jrest "joviandss-kubernetescsi/pkg/rest"
)


type SnapshotDesc struct {
	ld	LunDesc	// volume that this snapshot is made from
	name string	// name given by user
	sds string	// how snapshot is named inside joviandss
	id string	// that is sds without idFormat
	idFormat string
	// This id get formed by combining vds and sds encoded in base64 and separated by underscore
	csiID string	// this id provided to kubernetes
	isIntermidiateSnapshot bool // flag that indicate that this snapshot is intermediate snapshot to another volume
}

func IsSDS(vds string) bool {
	return vds[0] == 's'
}


func NewSnapshotDescFromName(lid LunDesc, name string) (*SnapshotDesc) {

	// Get universal volume ID
	var sd SnapshotDesc

	sd.ld = lid
	sd.name = name

	if len(name) <= 240 {
		if allowedSymbolsRegexp.MatchString(name) && len(name) < 240 {
			sd.sds = "sp_" + name
			sd.idFormat = "sp"
		} else if bname := jcom.JBase64FromStr(name); len(bname) <=240 {
			sd.sds = "sb_" + bname
			sd.idFormat = "sb"
		} else {
			preID := []byte(name)
			rawID := sha256.Sum256(preID)
			sd.sds	= fmt.Sprintf("ss_%X", rawID)
			sd.idFormat = "ss"
		}
	} else {
		preID := []byte(name)
		rawID := sha256.Sum256(preID)
		sd.sds	= fmt.Sprintf("ss_%X", rawID)
		sd.idFormat = "ss"
	}

	sd.csiID = fmt.Sprintf("%s_%s",
		sd.sds,
		base64.StdEncoding.EncodeToString([]byte(sd.ld.VDS())))
	return &sd
}

// parseSDS take sds string as 
func (sd *SnapshotDesc)parseSDS(sds string) (error) {

	parts := strings.Split(sds, "_")

	if len(parts) < 2 {
	 	return status.Error(codes.InvalidArgument, fmt.Sprintf("Snapshot descriptor have bad format %s", sds))
	}

	sd.sds = sds

	switch parts[0] {
	// Volume name in plain form
	case "sp":
		sd.idFormat = "sp"
		sd.name = strings.Join(parts[1:], "")
	// Volume name in form of base52
	case "sb":
		if name, err := jcom.JBase64ToStr(strings.Join(parts[1:], "")); err != nil {
			return err
		} else {
			sd.name = name
		}
		sd.idFormat = "sb"
	// Volume name, by default in sha512 hash
	case "s":
		sd.name = ""
		sd.idFormat = "s"
	// Volume name in form of sha512 hash
	case "ss":
		sd.name = ""
		sd.idFormat = "ss"
	default:
	 	return status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to identify type of snapshot naming %s", sds))
	}

	return nil
}

func NewSnapshotDescFromSDS(ld LunDesc, sds string) (*SnapshotDesc, error) {
	var sd SnapshotDesc

	sd.ld = ld

	if err := sd.parseSDS(sds); err != nil {
		return nil, err
	}

	sd.csiID = fmt.Sprintf("%s_%s",
		sd.sds,
		base64.StdEncoding.EncodeToString([]byte(sd.ld.VDS())))

	return &sd, nil
}

// NewSnapshotDescFromCSIID takes as argument csi snapshot id that is supplied to kubernetes and
// initialize desctiptor with it
func NewSnapshotDescFromCSIID(csiid string) (*SnapshotDesc, error) {
	var sd SnapshotDesc

	sd.csiID = csiid

	csiidl	:= strings.Split(csiid, "_")
	if len(csiidl) <= 2 {
		return nil, status.Errorf(codes.InvalidArgument, "Snapshot ID %s have bad format", csiid)
		// return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Unable to process snapshot token %s", csiid))
	}

	if vds, err := base64.StdEncoding.DecodeString(csiidl[len(csiidl)-1:][0]); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to decode volume section of snapshot ID %s have bad format, %s", csiid, err.Error())
	} else {
		if sd.ld, err = NewVolumeDescFromVDS(string(vds)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Volume section of snapshot ID %s have bad format, %s", csiid, err.Error())
		}
	}

	sd.sds = strings.Join(csiidl[:len(csiidl)-1], "_")

	sd.parseSDS(sd.sds)
	return &sd, nil
}

func (ps *SnapshotDesc)String () string {
	return fmt.Sprintf("%s_%s")
}

func (sd *SnapshotDesc)Name() string {

	if len(sd.name) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot name %+v", sd))
	}
	return sd.name
}

func (sd *SnapshotDesc)SDS() string {

	if len(sd.sds) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot descriptor string %+v", sd))
	}
	return sd.sds
}

func (sd *SnapshotDesc)CSIID() string {

	if len(sd.csiID) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot csi id %+v", sd))
	}
	return sd.csiID
}

func (sd *SnapshotDesc)GetVD() LunDesc {
	return sd.ld
}
