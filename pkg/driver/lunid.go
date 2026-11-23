package driver

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

type DataInstanceDesc interface {
	CSIID() string
}

type LunDesc interface {
	Name() string
	VDS() string
	CSIID() string
}

// type SnapshotId struct {
// 	name	string
// 	vds	string
// 	id	string
// }

const MaxVolumeNameLength int = 248

const allowedSymbolsPattern = `^[-\w]+$`

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

type VolumeDesc struct {
	name     string
	vds      string
	idFormat string
}

func NewVolumeDescFromName(name string) (*VolumeDesc, error) {

	// Get universal volume ID
	var vid VolumeDesc

	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}

	// The name on the appliance is also the CSI ID of the volume, and CSI
	// limits a string field to 128 bytes: with the three characters of the
	// prefix, a name of at most maxGeneratedName bytes fits, in its plain
	// form or in its base32 form. A longer one takes the hash form, which is
	// 67 characters. 128 is also within the 131 the appliance takes.
	if len(name) <= maxGeneratedName {
		if allowedSymbolsRegexp.MatchString(name) {
			vid.vds = "vp_" + name
			vid.idFormat = "vp"
		} else if bname := jcom.JBase32FromStr(name); len(bname) <= maxGeneratedName {
			vid.vds = "vb_" + bname
			vid.idFormat = "vb"
		} else {
			preID := []byte(name)
			rawID := sha256.Sum256(preID)
			vid.vds = fmt.Sprintf("vs_%X", rawID)
			vid.idFormat = "vs"
		}
	} else {
		preID := []byte(name)
		rawID := sha256.Sum256(preID)
		vid.vds = fmt.Sprintf("vs_%X", rawID)
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
		if name, err := jcom.JBase32ToStr(strings.Join(parts[1:], "")); err != nil {
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

// NewVolumeDescFromCSIID reads the ID of a volume that came in through a
// CSI request. The plugin does not control such an ID: it comes back from
// Kubernetes in every later request, and anyone allowed to create a static
// PersistentVolume can put any string there. So it is checked against the
// rules the appliance has for the name of a zvol or a NAS volume, and an ID
// that breaks them is refused with InvalidArgument before any REST call.
func NewVolumeDescFromCSIID(csiid string) (*VolumeDesc, error) {
	if err := checkApplianceName(csiid); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Volume ID %q is refused: %s", csiid, err.Error())
	}
	return NewVolumeDescFromVDS(csiid)
}

// The rules for the name of a zvol, a clone, a NAS volume and a NAS clone
// on the appliance. It refuses any other character, and a name of more than
// 132; 131 keeps one character of margin.
const (
	maxApplianceName = 131
	maxGeneratedName = 125 // the CSI name, or its base32 form, behind a prefix of three
)

var applianceNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// checkApplianceName says why a name cannot be the name of a volume on the
// appliance, and nil for one that can.
func checkApplianceName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("it is empty")
	case len(name) > maxApplianceName:
		return fmt.Errorf("it has %d characters, and a name on the appliance has at most %d", len(name), maxApplianceName)
	case !applianceNameRegexp.MatchString(name):
		return fmt.Errorf("a name on the appliance holds only a-z, A-Z, 0-9, _ and -")
	}
	return nil
}

func (vid *VolumeDesc) Name() string {

	if len(vid.name) == 0 {
		return vid.VDS()
	}
	return vid.name
}

func (vid *VolumeDesc) VDS() string {

	if len(vid.vds) == 0 {
		panic(fmt.Sprintf("Unable to identify volume sid %+v", vid))
	}
	return vid.vds
}

func (vid *VolumeDesc) CSIID() string {

	if len(vid.vds) == 0 {
		panic(fmt.Sprintf("Unable to identify volume sid and give proper CSIID %+v", vid))
	}
	return vid.vds
}
