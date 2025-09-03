package ports

import "context"

//go:generate mockgen -destination=./mocks/subscriber_ports_mock.go -package=mocks -source=subscriber.go

type Subscriber interface {
	Listen(ctx context.Context)
}
