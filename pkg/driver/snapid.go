package driver

import "fmt"

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


type phSnapIDs struct {
	id string
}

func NewPhysicalSnapshotId(from LunID, to LunID) (*phSnapIDs, error) {

	// Get universal volume ID
	var psid phSnapIDs

	psid.id = fmt.Sprintf("%s_%s", from.ID(), to.ID())
	return &psid, nil
}

