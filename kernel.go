package extract

import (
	"context"
)

var kernel = func() context.Context {
	ctx := context.Background()
	return ctx
}()
