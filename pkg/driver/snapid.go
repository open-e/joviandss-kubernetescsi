package driver

import (
	"fmt"
	"strings"
	"encoding/base64"
	"crypto/sha256"
)

// "crypto/sha256"
// "google.golang.org/grpc/codes"
// "google.golang.org/grpc/status"

// "fmt"
// "strings"

// type LunID interface {
// 	Name() string
// 	VID() string
// 	ID() string
// }


type PSID struct {
	lid LunID
	name string
	id string
	psid string
	csiID string
}

func NewPSIDFromName(lid LunID, name string) (*PSID) {

	// Get universal volume ID
	var psid PSID

	psid.lid = lid
	psid.name = name

	psid.id = strings.ToLower(fmt.Sprintf("%X", sha256.Sum256([]byte(name))))

	psid.psid = fmt.Sprintf("s_%s", psid.id)
	psid.csiID = fmt.Sprintf("%s_%s",
		base64.StdEncoding.EncodeToString([]byte(psid.lid.VID())),
		base64.StdEncoding.EncodeToString([]byte(psid.psid)))
	return &psid
}


func NewPSIDFromCSIID(csiid string) (*PSID) {
	var psid PSID
	
	csiidl := strings.Split(csiid, "_")
	base64.StdEncoding.DecodeString([]byte(psid.lid.VID()))
		base64.StdEncoding.DecodeString([]byte(psid.psid)))

	psid.lid = lid
	psid.name = name

	psid.id = strings.ToLower(fmt.Sprintf("%X", sha256.Sum256([]byte(name))))

	psid.psid = fmt.Sprintf("s_%s", psid.id)
	psid.csiID = fmt.Sprintf("%s_%s",
		base64.StdEncoding.EncodeToString([]byte(psid.lid.VID())),
		base64.StdEncoding.EncodeToString([]byte(psid.psid)))
	return &psid
}

func NewPSIDFromId(lid LunID, id string) (*PSID) {

	// Get universal volume ID
	var vid VolumeId

	if len(id) != 64 {
		return nil, status.Error(codes.InvalidArgument, "Incorrect snapshot ID") 
	}
	vid.name = ""
	vid.id = id
	vid.vid = "csi_s_" + vid.id
	return &vid, nil
}

func (ps *PSID)String () string {
	return fmt.Sprintf("%s_%s")
}
