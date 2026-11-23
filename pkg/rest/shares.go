package rest

import (
	"context"
)

func (s *RestEndpoint) CreateShare(ctx context.Context, desc *CreateShareDescriptor) RestError {
	path := apiPrefix + "/shares"
	return s.request(ctx, "POST", path, desc, createShareErrors, CodeCreated, CodeOK)
}
