package cmd

import (
	"github.com/go-chi/chi/v5"
	httphandler "trainstats-cancellations/http/handler"
)

type App struct {
	router  *chi.Mux
	handler httphandler.Handler
}

func NewApp(router *chi.Mux, handler httphandler.Handler) *App {
	return &App{router: router, handler: handler}
}

func (a *App) routes() {

}
