package shutdown

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/hexley21/fixup/pkg/logger"
)

// NotifyShutdown listens for OS signals (SIGINT, SIGTERM) to gracefully shut down the server.
// It logs the caught signal, attempts to close the server, and sends a signal to the shutdownError channel
// indicating whether the shutdown was successful or not.
func NotifyShutdown(serverCloser io.Closer, logger logger.Logger, shutdownError chan<- struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("caught signal", "signal", (<-sig).String())

	if err := serverCloser.Close(); err != nil {
		shutdownError <- struct{}{}
		return
	}

	shutdownError <- struct{}{}
}
