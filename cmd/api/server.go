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
		Addr:         ":8081",
		Handler:      a.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	a.routes()

	serverErrors := make(chan error, 1)

	go func() {
		log.Println("starting api...")
		serverErrors <- srv.ListenAndServe()
	}()

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		log.Printf("caught signal %s", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		close(shutdownError)
	}()

	select {
	case err := <-serverErrors:
		return err
	case err := <-shutdownError:
		return err
	}
}
