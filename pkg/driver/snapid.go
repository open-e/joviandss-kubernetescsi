package driver

import (
	"fmt"
	"strings"
	"encoding/base64"
	"crypto/sha256"
	
	jrest "joviandss-kubernetescsi/pkg/rest"
)


type SnapshotDesc struct {
	lid LunID	// volume that this snapshot is made from
	name string	// name given by user
	sds string	// how snapshot is named inside joviandss
	idFormat string
	csiID string	// this id getsend to kubernetes
}

func NewSnapshotDescFromName(lid LunID, name string) (*SnapshotDesc) {

	// Get universal volume ID
	var sd SnapshotDesc

	psid.lid = lid
	psid.name = name
	
	if len(name) <= 240 {
		if allowedSymbolsRegexp.MatchString(name) {
			sd.vds = "sp_" + name
			sd.idFormat = "sp"
		} else if bname, err := jcom.JBase64FromSrt(name); len(bname) <=240 {
			if err != nil {
				return nil, err
			}
			sd.vds = "sb_" + bname
			sd.idFormat = "sb"
		} else {
			preID := []byte(name)
			rawID := sha256.Sum256(preID)
			sd.vds	= fmt.Sprintf("ss_%X", rawID)
			sd.idFormat = "ss"
		}
	} else {
		preID := []byte(name)
		rawID := sha256.Sum256(preID)
		vid.vds	= fmt.Sprintf("ss_%X", rawID)
		vid.idFormat = "ss"
	}

	psid.id = strings.ToLower(fmt.Sprintf("%X", sha256.Sum256([]byte(name))))

	psid.psid = fmt.Sprintf("s_%s", psid.id)
	psid.csiID = fmt.Sprintf("%s_%s",
		base64.StdEncoding.EncodeToString([]byte(psid.lid.VDS())),
		base64.StdEncoding.EncodeToString([]byte(psid.psid)))
	return &psid
}


func NewSnapshotDescFromCSIID(csiid string) (*SnapshotDesc, jrest.RestError) {
	var psid SnapshotDesc
	
	csiidl	:= strings.Split(csiid, "_")
	if len(csiidl) != 2 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Unable to process snapshot token %s", csiid))
	}

	if vid, err := base64.StdEncoding.DecodeString(csiidl[0]); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Unable to process snapshot token %s, decoding failed %s", csiid, err.Error()))
	} else {
		if psid.lid, err = NewVolumeIdFromId(string(vid)); err != nil {
			return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("Unable to restore voprocess snapshot token %s, decoding failed %s", csiid, err.Error())) 
		}
	}
	if err != nil {
		return nil
	}
	psid	:= base64.StdEncoding.DecodeString(csiidl[1])

	psid.lid = lid
	psid.name = name

	psid.id = strings.ToLower(fmt.Sprintf("%X", sha256.Sum256([]byte(name))))

	psid.psid = fmt.Sprintf("s_%s", psid.id)
	psid.csiID = fmt.Sprintf("%s_%s",
		base64.StdEncoding.EncodeToString([]byte(psid.lid.VID())),
		base64.StdEncoding.EncodeToString([]byte(psid.psid)))
	return &psid
}

// func NewSnapshotDescFromId(lid LunID, id string) (*SnapshotDesc) {
// 
// 	// Get universal volume ID
// 	var vid VolumeId
// 
// 	if len(id) != 64 {
// 		return nil, status.Error(codes.InvalidArgument, "Incorrect snapshot ID") 
// 	}
// 	vid.name = ""
// 	vid.id = id
// 	vid.vid = "csi_s_" + vid.id
// 	return &vid, nil
// }

func (ps *SnapshotDesc)String () string {
	return fmt.Sprintf("%s_%s")
}
