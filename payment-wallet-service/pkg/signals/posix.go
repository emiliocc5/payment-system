package signals

import (
	"os"
	"syscall"
)

var _shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}
