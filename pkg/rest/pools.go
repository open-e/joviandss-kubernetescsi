package rest

import (
	"context"
	"fmt"
)

func (s *RestEndpoint) GetPool(ctx context.Context, pool string) (*ResourcePool, RestError) {
	path := fmt.Sprintf(apiPrefix+"/pools/%s", pool)
	return readObject[ResourcePool](ctx, s, path, getPoolErrors, CodeOK)
}
