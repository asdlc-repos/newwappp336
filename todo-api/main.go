package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/asdlc-repos/newwappp336/todo-api/internal/handlers"
	"github.com/asdlc-repos/newwappp336/todo-api/internal/middleware"
	"github.com/asdlc-repos/newwappp336/todo-api/internal/store"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	s := store.New()
	api := handlers.New(s)

	mux := http.NewServeMux()
	api.Register(mux)

	handler := middleware.Chain(mux, middleware.Logging, middleware.CORS)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("todo-api listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
