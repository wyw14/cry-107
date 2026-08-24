package main

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

	"github.com/wyw14/cry-107/internal/api"
)

func main() {
	address := os.Getenv("KILNGUARD_ADDR")
	if address == "" {
		address = "127.0.0.1:21207"
	}
	dataDirectory := os.Getenv("KILNGUARD_DATA")
	if dataDirectory == "" {
		dataDirectory = "data"
	}
	system, err := api.NewSystem(dataDirectory)
	if err != nil {
		log.Fatalf("initialize KilnGuard: %v", err)
	}
	server := &http.Server{
		Addr: address, Handler: api.Router(system),
		ReadHeaderTimeout: 3 * time.Second, IdleTimeout: 30 * time.Second,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown KilnGuard: %v", err)
		}
	}()
	fmt.Printf("KilnGuard listening on http://%s\n", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve KilnGuard: %v", err)
	}
}
