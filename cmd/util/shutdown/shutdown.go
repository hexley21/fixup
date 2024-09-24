package shutdown

import (
	"github.com/labstack/echo/v4"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func NotifyShutdown(serverCloser io.Closer, logger echo.Logger, shutdownError chan<- error) {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

logger.Info("caught signal", "signal", (<-sig).String())

	if err := serverCloser.Close(); err != nil {
		shutdownError <- err
		return
	}

	shutdownError <- nil
}
