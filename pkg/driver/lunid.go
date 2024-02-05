package driver

import (
	"crypto/sha256"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"fmt"
	"strings"
)

type LunID interface {
	Name() string 
	VID() string
	ID() string 
}

type SnapshotId struct {
	name	string
	vid	string
	id	string
}

func NewSnapshotIdFromName(name string) (*SnapshotId, error) {

	// Get universal volume ID
	var sid SnapshotId

	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request") 
	}

	sid.name = name
	preID := []byte(name)
	rawID := sha256.Sum256(preID)
	id := strings.ToLower(fmt.Sprintf("%X", rawID))
	sid.id = id
	sid.vid = "csi_s_" + sid.id
	return &sid, nil
}

func NewSnapshotIdFromId(id string) (*SnapshotId, error) {

	// Get universal volume ID
	var sid SnapshotId

	if len(id) != 64 {
		return nil, status.Error(codes.InvalidArgument, "Incorrect snapshot ID") 
	}
	sid.name = ""
	sid.id = id
	sid.vid = "csi_s_" + sid.id
	return &sid, nil
}

func (vid *SnapshotId)Name() string {

	if len(vid.name) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot name %+v", vid))
	}
	return vid.name
}

func (vid *SnapshotId)VID() string {

	if len(vid.vid) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot sid %+v", vid))
	}
	return vid.vid
}

func (vid *SnapshotId)ID() string {

	if len(vid.id) == 0 {
	 	panic(fmt.Sprintf("Unable to identify snapshot id %+v", vid))
	}
	return vid.id
}


type VolumeId struct {
	name	string
	vid	string
	id	string
}

func NewVolumeIdFromName(name string) (*VolumeId, error) {

	// Get universal volume ID
	var vid VolumeId

	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request") 
	}

	vid.name = name
	preID := []byte(name)
	rawID := sha256.Sum256(preID)
	id := strings.ToLower(fmt.Sprintf("%X", rawID))
	vid.id = id
	vid.vid = "csi_v_" + vid.id
	return &vid, nil
}

func NewVolumeIdFromId(id string) (*VolumeId, error) {

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

func (vid *VolumeId)Name() string {

	if len(vid.name) == 0 {
	 	panic(fmt.Sprintf("Unable to identify volume name %+v", vid))
	}
	return vid.name
}

func (vid *VolumeId)VID() string {

	if len(vid.vid) == 0 {
	 	panic(fmt.Sprintf("Unable to identify volume sid %+v", vid))
	}
	return vid.vid
}

func (vid *VolumeId)ID() string {

	if len(vid.id) == 0 {
	 	panic(fmt.Sprintf("Unable to identify volume id %+v", vid))
	}
	return vid.id
}


