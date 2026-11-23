package rest

import (
	"context"
)

// GetNASVolumeSnapshot provides information about specific NAS volume snapshot requested
func (s *RestEndpoint) GetNASVolumeSnapshot(ctx context.Context, pool string, vname string, sname string) (sdp *ResourceNASVolumeSnapshot, err RestError) {
	path := apiPath("/pools/%s/nas-volumes/%s/snapshots/%s", pool, vname, sname)
	return readObject[ResourceNASVolumeSnapshot](ctx, s, path, getNASVolumeSnapshotErrors, CodeOK, CodeAccepted)
}
