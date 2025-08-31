package ports

import "context"

type Subscriber interface {
	Listen(ctx context.Context)
}
