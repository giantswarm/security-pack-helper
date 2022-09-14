package reportchangerequestcleaner

import "context"

type Interface interface {
	xCheck(ctx context.Context) bool
}
