package utils

import (
	"time"
	"net/http"
	"strings"
	"raspberry_sensors/internal/api"
	"raspberry_sensors/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

func QueryWithRetry(url string, timeout time.Duration) error {
    // Deadline is the maximum time we allow for retries
    deadline := time.Now().Add(timeout)
    
    var lastErr error

    for time.Now().Before(deadline) {
        resp, err := http.Get(url)
        if err != nil {
            // Check if the error contains "connection refused"
            if strings.Contains(err.Error(), "connection refused") {
                lastErr = err
                time.Sleep(100 * time.Millisecond) // Short delay before retrying
                continue
            }
            return err // If it's a different error, return immediately
        }
        resp.Body.Close()
        return nil // Successful
    }
    // If we exit the loop, it means we exhausted the retries
    return lastErr
}

func WaitForExitSignal(server *api.Server) {
	// Create a channel to receive OS signals
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	defer close(server.QuitChan)

	select {
	case <-exitChan:
        logger.GlobalLogger.Info("Programm stopped. Shutting down...")
			resp, _ := http.Get("http://localhost:8080/sensors/stop")
			defer resp.Body.Close()
			logger.GlobalLogger.Info("Bye!")
	case <-server.QuitChan:
		// Nothing else to do, just acknowledge the channel
	}
}