package http_utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServer(port int, handler http.Handler) {
	svc := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	go func() {
		log.Printf("Starting server on port %d...\n", port)
		err := svc.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on port %d: %v", port, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Println("Waiting for shutdown signal...")

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := svc.Shutdown(ctx)
	if err != nil {
		log.Fatalf("could not shutdown server: %v", err)
	}
	log.Println("Server gracefully stopped")
}
