package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *App) serve() error {
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      a.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Set up the routes
	a.routes()

	// Channel to receive any errors returned by the ListenAndServe
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Println("starting api...")
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to receive shutdown errors
	shutdownError := make(chan error)

	// Set up signal handling
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		log.Printf("caught signal %s", s.String())

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Trigger graceful shutdown
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		close(shutdownError)
	}()

	// Block until an error is received or the server is shut down
	select {
	case err := <-serverErrors:
		return err
	case err := <-shutdownError:
		return err
	}
}
