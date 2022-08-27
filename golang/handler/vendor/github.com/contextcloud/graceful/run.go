package graceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/contextcloud/graceful/srv"
)

func Run(ctx context.Context, s srv.Startable) {
	serverCtx, serverStopCtx := context.WithCancel(ctx)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Print(err)
		}

		serverStopCtx()
	}()

	// start it.
	if err := s.Start(serverCtx); err != nil {
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
