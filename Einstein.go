package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Einstein/neurons"
)

const (
	serverAddress        = "127.0.0.1:80"
	mattermostAPIVersion = 4
)

func main() {
	log.Printf("Starting server at %s and listening...\n", serverAddress)

	neurons.Greet()

	route := http.NewServeMux()

	// Handle registers functions that already implement the Handler interface, meaning that their ServeHTTP method will be called by the server multiplexer to handle requests
	// HandleFunc registers functions of the signature func(ResponseWriter, *Request) directly
	// HandlerFunc converts functions of the signature func(ResponseWriter, *Request) to a function that implements the Handler interface
	route.HandleFunc(
		fmt.Sprintf("/api/v%d/plugins/webapp", mattermostAPIVersion),
		neurons.Gzip(neurons.SetAPIHeaders(neurons.GetPulse)),
	)
	route.HandleFunc(
		fmt.Sprintf("/api/v%d/users/status/ids", mattermostAPIVersion),
		neurons.Gzip(neurons.SetAPIHeaders(neurons.GetJob)),
	)
	route.HandleFunc(
		fmt.Sprintf("/api/v%d/users/ids", mattermostAPIVersion),
		neurons.SetAPIHeaders(neurons.GetResponse),
	)
	route.HandleFunc(
		fmt.Sprintf("/api/v%d/", mattermostAPIVersion),
		neurons.Gzip(neurons.FakeAPINotFound), // for some reason API headers won't be set in this case
	)
	route.HandleFunc("/api/", neurons.Gzip(neurons.SetAPIHeaders(neurons.SetStaticHeaders(neurons.FakeAPINotFound))))
	route.HandleFunc("/static/", neurons.Gzip(neurons.SendFile))

	// matches all paths not matched by the above registered patterns
	route.HandleFunc("/", neurons.Gzip(neurons.SetAPIHeaders(neurons.SetStaticHeaders(neurons.FakeNotFound)))) // this one has API headers for some reason

	// creating a Server this way allows for more potential customization
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: neurons.SetCommonHeaders(route),
	}

	exit := make(chan struct{}, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// ListenAndServe is naturally blocking
		// ListenAndServe can also instead take in (address, handler) so something like http.ListenAndServe(serverAddress, neurons.SetCommonHeaders(route))
		// if handler is nil then Go will automatically use http.DefaultServeMux
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go neurons.GetInput(exit)

	select {
	case <-interrupt:
		fmt.Println()
	case <-exit:
	}

	log.Println("Server stopped.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// do stuff necessary here?
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}
	log.Println("Server exited properly.")

}
