package rest

import (
	"context"
)

func (s *RestEndpoint) GetPool(ctx context.Context, pool string) (*ResourcePool, RestError) {
	path := apiPath("/pools/%s", pool)
	return readObject[ResourcePool](ctx, s, path, getPoolErrors, CodeOK)
}
