package rest

import (
	"context"
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////
// How a call ends
//
// Every call ends in its success or in failedReply (errors.go) with the
// catalogue of the errors that call can get (error_catalogue.go). The
// status of a reply is looked at first; a body is decoded only once the
// status says which type it holds.

func isOneOf(status int, ok []int) bool {
	for _, s := range ok {
		if status == s {
			return true
		}
	}
	return false
}

// request sends a request whose reply holds nothing the plugin reads. A
// status of ok is its success. The three functions after it send a GET,
// through getWithData: the one way a GET of the plugin is sent.
func (s *RestEndpoint) request(ctx context.Context, method, path string, data interface{},
	entries []errorEntry, ok ...int,
) RestError {
	status, body, err := s.rp.Send(ctx, method, path, data, ok[0])
	if err == nil && isOneOf(status, ok) {
		return nil
	}
	return s.failedReply(ctx, method, path, entries, status, body, err)
}

// readObject sends a GET and decodes the data of its reply into a new T.
func readObject[T any, PT interface {
	*T
	replyType
}](ctx context.Context, s *RestEndpoint, path string, entries []errorEntry, ok ...int) (*T, RestError) {
	status, body, err := s.rp.getWithData(ctx, path)
	if err == nil && isOneOf(status, ok) {
		v := new(T)
		if err = decodeData(body, PT(v)); err == nil {
			return v, nil
		}
	}
	return nil, s.failedReply(ctx, "GET", path, entries, status, body, err)
}

// readEntries sends the GET of one page of a list and decodes its entries.
func readEntries[T any, PT interface {
	*T
	replyType
}](ctx context.Context, s *RestEndpoint, path string, entries []errorEntry) (*ResultEntries, RestError) {
	status, body, err := s.rp.getWithData(ctx, path)
	if err == nil && (status == CodeOK || status == CodeCreated) {
		var page []T
		var results int64
		if page, results, err = decodeEntries[T, PT](body); err == nil {
			return &ResultEntries{Results: results, Entries: &page}, nil
		}
	}
	return nil, s.failedReply(ctx, "GET", path, entries, status, body, err)
}

// readItems sends a GET whose reply is a plain list and decodes its items.
func readItems[T any, PT interface {
	*T
	replyType
}](ctx context.Context, s *RestEndpoint, path string, entries []errorEntry) ([]T, RestError) {
	status, body, err := s.rp.getWithData(ctx, path)
	if err == nil && status == CodeOK {
		var items []T
		if items, err = decodeItems[T, PT](body); err == nil {
			return items, nil
		}
	}
	return nil, s.failedReply(ctx, "GET", path, entries, status, body, err)
}

///////////////////////////////////////////////////////////////////////////////
// Volumes

func (s *RestEndpoint) GetVolume(ctx context.Context, pool string, vname string) (*ResourceVolume, RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s", pool, vname)
	return readObject[ResourceVolume](ctx, s, path, getVolumeErrors, CodeOK)
}

func (s *RestEndpoint) CreateVolume(ctx context.Context, pool string, vol CreateVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes", pool)
	return s.request(ctx, "POST", path, vol, createVolumeErrors, CodeCreated, CodeOK)
}

// DeleteVolume deletes a volume. The appliance refuses the delete of a
// volume that has snapshots or clones.
func (s *RestEndpoint) DeleteVolume(ctx context.Context, pool string, vname string, data DeleteVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s", pool, vname)
	return s.request(ctx, "DELETE", path, data, deleteVolumeErrors, CodeNoContent, CodeOK)
}

// GetVolumeSnapshot provides information about specific volume snapshot requested
func (s *RestEndpoint) GetVolumeSnapshot(ctx context.Context, pool string, vname string, sname string) (sdp *ResourceSnapshot, err RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots/%s", pool, vname, sname)
	return readObject[ResourceSnapshot](ctx, s, path, getVolumeSnapshotErrors, CodeOK, CodeAccepted)
}

// # Create Snapshot from existing volume
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *vid*
//   - *vid* physical volume id as it used by JovianDSS
//   - *desc* data, including snapahot name, that would be transfered to create snapshot
func (s *RestEndpoint) CreateSnapshot(ctx context.Context, pool string, vid string, desc *CreateSnapshotDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots", pool, vid)
	return s.request(ctx, "POST", path, desc, createSnapshotErrors, CodeOK, CodeCreated)
}

func (s *RestEndpoint) DeleteSnapshot(ctx context.Context, pool string, vname string, sname string, data DeleteSnapshotDescriptor) (err RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots/%s", pool, vname, sname)
	return s.request(ctx, "DELETE", path, data, deleteSnapshotErrors, CodeNoContent)
}
