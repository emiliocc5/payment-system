package ports

import "context"

//go:generate mockgen -destination=./mocks/consumer_ports_mock.go -package=mocks -source=consumer.go

// TODO check if this file is necessary
type Consumer interface {
	Listen(ctx context.Context)
}
