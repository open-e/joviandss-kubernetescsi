package rest

import (
	"context"
	"fmt"
)

func (s *RestEndpoint) GetNASVolume(ctx context.Context, pool string, vname string) (*ResourceNASVolume, RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s", pool, vname)
	return readObject[ResourceNASVolume](ctx, s, path, getNASVolumeErrors, CodeOK)
}

func (s *RestEndpoint) CreateNASVolume(ctx context.Context, pool string, desc *CreateNASVolumeDescriptor) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes", pool)
	return s.request(ctx, "POST", path, desc, createNASVolumeErrors, CodeCreated, CodeOK)
}

func (s *RestEndpoint) DeleteNASVolume(ctx context.Context, pool string, nvds string) RestError {
	path := fmt.Sprintf(apiPrefix+"/pools/%s/nas-volumes/%s", pool, nvds)
	return s.request(ctx, "DELETE", path, nil, deleteNASVolumeErrors, CodeNoContent)
}
