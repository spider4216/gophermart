package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/gophermart/internal/handler"
	"github.com/spider4216/gophermart/internal/middleware"
	"github.com/spider4216/gophermart/internal/repository"
	"github.com/spider4216/gophermart/internal/service"
)

func main() {
	app := newApp()

	if err := app.Run(); err != nil {
		log.Fatal("Cannot run app", err)
	}

	app.logger.Debug("Config: ", app.cfg)

	repo := repository.New(app.store)
	service := service.New(repo, app.logger, app.cfg)
	handler := handler.New(app.cfg, app.logger, service)
	middlewares := middleware.New(app.logger, app.cfg, service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.WithLogging)
		r.Use(middlewares.WithGzip)

		r.Get("/ping", http.HandlerFunc(handler.Ping))
		r.Post("/api/user/register", http.HandlerFunc(handler.SignUp))
		r.Post("/api/user/login", http.HandlerFunc(handler.SignIn))
		r.With(middlewares.WithJwt).Post("/api/user/orders", http.HandlerFunc(handler.RegOrder))
		r.With(middlewares.WithJwt).Get("/api/user/orders", http.HandlerFunc(handler.GetUserOrders))
	})

	srv := &http.Server{
		Addr:         app.cfg.RunAddress,
		Handler:      r,
		ReadTimeout:  app.cfg.ReadTimeout,
		WriteTimeout: app.cfg.WriteTimeout,
		IdleTimeout:  app.cfg.IdleTimeout,
	}

	log.Printf("Listen on: %s", app.cfg.RunAddress)

	if err := srv.ListenAndServe(); err != nil {
		app.logger.Fatalf("Server error: %s", err)
	}

	app.logger.Infof("Starting server on %s", app.cfg.RunAddress)
}
