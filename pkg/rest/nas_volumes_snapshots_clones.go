package rest

import (
	"context"
	"fmt"
)

func (s *RestEndpoint) GetNASVolumeSnapshotClones(ctx context.Context, pool string, vds string, sds string) (clones []ResourceVolumeSnapshotClones, err RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s/snapshots/%s/clones", pool, vds, sds)
	return readItems[ResourceVolumeSnapshotClones](ctx, s, path, getNASVolumeSnapshotClonesErrors)
}

// # CreateNASClone creates clone of a NAS volume from snapshot
//
// Arguments:
//
//   - *ctx* current request context is expected to be with logger object at loggerKey
//   - *pool* name of JovianDSS Pool that is storing volume described with *nvid*
//   - *nvid* physical volume id as it used by JovianDSS
//   - *nsds* the snapshot the clone is made of
//   - *desc* data, including new volume name, that would be transfered to create clone
func (s *RestEndpoint) CreateNASClone(ctx context.Context, pool string, nvid string, nsds string, desc CloneNASVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s/snapshots/%s/clones", pool, nvid, nsds)
	return s.request(ctx, "POST", path, desc, createNASCloneErrors, CodeOK, CodeCreated)
}

func (s *RestEndpoint) DeleteNASClone(ctx context.Context, pool string, nvds string, nsds string, ncds string) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s/snapshots/%s/clones/%s", pool, nvds, nsds, ncds)
	return s.request(ctx, "DELETE", path, nil, deleteNASCloneErrors, CodeNoContent)
}
