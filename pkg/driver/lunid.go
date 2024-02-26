package driver

import (
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"crypto/sha256"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "joviandss-kubernetescsi/pkg/common"
)

type LunDesc interface {
	Name() string 
	VDS() string
}

// type SnapshotId struct {
// 	name	string
// 	vds	string
// 	id	string
// }

const MaxVolumeNameLength int = 248

const allowedSymbolsPattern = "^[abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_]+$"
var allowedSymbolsRegexp = regexp.MustCompile(allowedSymbolsPattern)

func nameToID(name string) string {

    	// Replace each non-allowed symbol with its hexadecimal representation
    	transformedString := allowedSymbolsRegexp.ReplaceAllStringFunc(name, func(s string) string {
    	    // Convert the non-allowed symbol to its hexadecimal representation
    	    runeValue := []rune(s)[0]
    	    hexRepresentation := strconv.FormatInt(int64(runeValue), 16)
    	    return "_" + hexRepresentation
    	})
	return transformedString
}

// func NewSnapshotIdFromName(name string) (*SnapshotId, error) {
// 
// 	// Get universal volume ID
// 	var sid SnapshotId
// 
// 	if len(name) == 0 {
// 		return nil, status.Error(codes.InvalidArgument, "Name missing in request") 
// 	}
// 
// 	if len(name) <= 240 {
// 		if  allowedSymbolsRegexp.MatchString(name) {
// 			sid.vds = "sp_" + name
// 		}
// 	}
// 	sid.name = name
// 	preID := []byte(name)
// 	rawID := sha256.Sum256(preID)
// 	id := strings.ToLower(fmt.Sprintf("%X", rawID))
// 	sid.id = id
// 
// 	return &sid, nil
// }
// 
// func NewSnapshotIdFromId(id string) (*SnapshotId, error) {
// 
// 	// Get universal volume ID
// 	var sid SnapshotId
// 
// 	if len(id) != 64 {
// 		return nil, status.Error(codes.InvalidArgument, "Incorrect snapshot ID") 
// 	}
// 	sid.name = ""
// 	sid.id = id
// 	sid.vid = "csi_s_" + sid.id
// 	return &sid, nil
// }

//func (vid *SnapshotId)Name() string {
//
//	if len(vid.name) == 0 {
//	 	panic(fmt.Sprintf("Unable to identify snapshot name %+v", vid))
//	}
//	return vid.name
//}
//
//func (vid *SnapshotId)VID() string {
//
//	if len(vid.vid) == 0 {
//	 	panic(fmt.Sprintf("Unable to identify snapshot sid %+v", vid))
//	}
//	return vid.vid
//}
//
//func (vid *SnapshotId)ID() string {
//
//	if len(vid.id) == 0 {
//	 	panic(fmt.Sprintf("Unable to identify snapshot id %+v", vid))
//	}
//	return vid.id
//}


type VolumeDesc struct {
	name		string
	vds		string
	idFormat	string
}

func NewVolumeDescFromName(name string) (*VolumeDesc, error) {

	// Get universal volume ID
	var vid VolumeDesc

	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}

	if len(name) <= 240 {
		if allowedSymbolsRegexp.MatchString(name) {
			vid.vds = "vp_" + name
			vid.idFormat = "vp"
		} else if bname := jcom.JBase64FromStr(name); len(bname) <=240 {
			vid.vds = "vb_" + bname
			vid.idFormat = "vb"
		} else {
			preID := []byte(name)
			rawID := sha256.Sum256(preID)
			vid.vds	= fmt.Sprintf("vs_%X", rawID)
			vid.idFormat = "vs"
		}
	} else {
		preID := []byte(name)
		rawID := sha256.Sum256(preID)
		vid.vds	= fmt.Sprintf("vs_%X", rawID)
		vid.idFormat = "vs"
	}

	vid.name = name
	return &vid, nil
}

// func NewVolumeDescFromId(id string) (*VolumeDesc, error) {
// 
// 	// Get universal volume ID
// 	var vid VolumeDesc
// 
// 	if len(id) != 64 {
// 		return nil, status.Error(codes.InvalidArgument, "Incorrect snapshot ID") 
// 	}
// 	vid.name = ""
// 	vid.id = id
// 	vid.vid = "csi_s_" + vid.id
// 	return &vid, nil
// }


func IsVDS(vds string) bool {
	return vds[0] == 'v'
}

func NewVolumeDescFromVDS(vds string) (*VolumeDesc, error) {

	// Get universal volume ID
	var vd VolumeDesc

	parts := strings.Split(vds, "_")
	if len(parts) < 2 {
	 	return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Volume descriptor have bad format %s", vds))
	}
	
	vd.vds = vds

	switch parts[0] {
	// Volume name in plain form
	case "vp":
		vd.idFormat = "vp"
		vd.name = strings.Join(parts[1:], "")
	// Volume name in form of base52
	case "vb":
		if name, err := jcom.JBase64ToStr(strings.Join(parts[1:], "")); err != nil {
			return nil, err
		} else {
			vd.name = name
		}
		vd.idFormat = "vb"
	// Volume name, by default in sha512 hash
	case "v":
		vd.name = ""
		vd.idFormat = "v"
	// Volume name in form of sha512 hash
	case "vs":
		vd.name = ""
		vd.idFormat = "vs"
	default:
	 	return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to identify type of volume naming %s", vds))
	}

	return &vd, nil
}

func (vid *VolumeDesc)Name() string {

	if len(vid.name) == 0 {
		return vid.VDS()
	}
	return vid.name
}

func (vid *VolumeDesc)VDS() string {

	if len(vid.vds) == 0 {
	 	panic(fmt.Sprintf("Unable to identify volume sid %+v", vid))
	}
	return vid.vds
}
