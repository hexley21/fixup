package shutdown

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/hexley21/fixup/pkg/logger"
)

func NotifyShutdown(serverCloser io.Closer, logger logger.Logger, shutdownError chan<- error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("caught signal", "signal", (<-sig).String())

	if err := serverCloser.Close(); err != nil {
		shutdownError <- err
		return
	}

	shutdownError <- nil
}
