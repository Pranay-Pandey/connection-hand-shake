package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(closers ...interface{ Close() error }) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			// Consider using a proper logging package here
			println("Error closing:", err.Error())
		}
	}
}
