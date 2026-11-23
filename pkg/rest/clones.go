package rest

import (
	"context"
	"fmt"
)

func (s *RestEndpoint) GetVolumeSnapshotClones(ctx context.Context, pool string, vds string, sds string) (clones []ResourceVolumeSnapshotClones, err RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots/%s/clones", pool, vds, sds)
	return readItems[ResourceVolumeSnapshotClones](ctx, s, path, getVolumeSnapshotClonesErrors)
}

// # CreateClone creates clone of a volume from snapshot
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *vid*
//   - *vid* physical volume id as it used by JovianDSS
//   - *desc* data, including new volume name, that would be transfered to create clone
func (s *RestEndpoint) CreateClone(ctx context.Context, pool string, vid string, desc CloneVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/clone", pool, vid)
	return s.request(ctx, "POST", path, desc, createCloneErrors, CodeOK, CodeCreated)
}

func (s *RestEndpoint) DeleteClone(ctx context.Context, pool string, vds string, sds string, cds string, desc DeleteVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/volumes/%s/snapshots/%s/clones/%s", pool, vds, sds, cds)
	return s.request(ctx, "DELETE", path, desc, deleteCloneErrors, CodeNoContent)
}
