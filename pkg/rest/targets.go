package rest

import (
	"context"
	"fmt"
)

func (s *RestEndpoint) GetTarget(ctx context.Context, pool string, tname string) (*ResourceTarget, RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s", pool, tname)
	return readObject[ResourceTarget](ctx, s, path, getTargetErrors, CodeOK)
}

// CreateTarget creates a target. A status of 202 is not its success: no
// recording shows one, and the reply would say nothing about the target.
func (s *RestEndpoint) CreateTarget(ctx context.Context, pool string, desc *CreateTargetDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets", pool)
	return s.request(ctx, "POST", path, desc, createTargetErrors, CodeCreated, CodeOK)
}

// DeleteTarget deletes a target. A target that is gone already is an error
// of its own, RestErrorResourceDNETarget, and so is a missing pool.
func (s *RestEndpoint) DeleteTarget(ctx context.Context, pool string, tname string) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s", pool, tname)
	return s.request(ctx, "DELETE", path, nil, deleteTargetErrors, CodeNoContent)
}

func (s *RestEndpoint) AttachVolumeToTarget(ctx context.Context, pool string, tname string, desc *TargetLunDescriptor) (err RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s/luns", pool, tname)
	return s.request(ctx, "POST", path, desc, attachVolumeToTargetErrors, CodeCreated, CodeOK)
}

// DettachVolumeFromTarget removes a volume from a target. Its three replies
// of 404 are three errors: the pool is missing (RestErrorResourceDNEPool),
// the target is missing (RestErrorResourceDNETarget), and the volume is not
// attached to the target, which is still there (RestErrorResourceDNELun).
func (s *RestEndpoint) DettachVolumeFromTarget(ctx context.Context, pool string, tname string, vname string) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/san/iscsi/targets/%s/luns/%s", pool, tname, vname)
	return s.request(ctx, "DELETE", path, nil, dettachVolumeFromTargetErrors, CodeNoContent)
}
