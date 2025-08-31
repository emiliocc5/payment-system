package signals

import (
	"os"
	"os/signal"
)

var _onlyOneSignalHandler = make(chan struct{})

func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(_onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, _shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}
