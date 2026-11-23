package rest

import (
	"context"
)

func (s *RestEndpoint) CreateShare(ctx context.Context, desc *CreateShareDescriptor) RestError {
	path := apiPath("/shares")
	return s.request(ctx, "POST", path, desc, createShareErrors, CodeCreated, CodeOK)
}

// GetShare reads a share by its name. A share that does not exist is
// RestErrorResourceDNEShare.
func (s *RestEndpoint) GetShare(ctx context.Context, name string) (*ResourceShare, RestError) {
	path := apiPath("/shares/%s", name)
	return readObject[ResourceShare](ctx, s, path, getShareErrors, CodeOK)
}
