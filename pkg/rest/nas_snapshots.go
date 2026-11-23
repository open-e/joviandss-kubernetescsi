package rest

import (
	"context"
)

// # Create Snapshot from existing NAS Volume
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing nas volume described with *vid*
//   - *vid* physical nas volume id as it used by JovianDSS
//   - *desc* data, including snapahot name, that would be transfered to create snapshot
func (s *RestEndpoint) CreateNASSnapshot(ctx context.Context, pool string, nvid string, desc *CreateNASSnapshotDescriptor) RestError {
	path := apiPath("/pools/%s/nas-volumes/%s/snapshots", pool, nvid)
	return s.request(ctx, "POST", path, desc, createNASSnapshotErrors, CodeCreated, CodeOK)
}

// DeleteNASSnapshot deletes a snapshot of a NAS volume. The appliance does
// not refuse the delete of a NAS snapshot that has a clone: it deletes the
// clone with it. A caller reads the clones first.
func (s *RestEndpoint) DeleteNASSnapshot(ctx context.Context, pool string, nvid string, sname string) (err RestError) {
	path := apiPath("/pools/%s/nas-volumes/%s/snapshots/%s", pool, nvid, sname)
	return s.request(ctx, "DELETE", path, nil, deleteNASSnapshotErrors, CodeNoContent)
}
