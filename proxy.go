package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var (
	// ServerPort is the default port on which the HTTP server listens
	ServerPort = "8000"

	// ServerTimeout is the default amount of time for which the proxy server waits for the
	// HTTP client to finish
	ServerTimeout = time.Duration(5 * time.Second)
)

// proxyServerHTTPHandler implements the same handler for
// GET and POST requests to the /proxy/ endpoint
func proxyServerHTTPHandler(w http.ResponseWriter, r *http.Request) {
	receivedURL := mux.Vars(r)["url"]
	log.Printf("Received %s request for %s", r.Method, receivedURL)

	if len(receivedURL) == 0 {
		errStr := "No url specified in path"
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, errStr)
		return
	}

	req, err := http.NewRequest(r.Method, receivedURL, r.Body)
	// need to specify the same User-Agent as the incoming request
	req.Header.Set("User-Agent", r.UserAgent())

	ctx, cancel := context.WithTimeout(req.Context(), ServerTimeout)
	defer cancel()

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		io.WriteString(w, err.Error())
		return
	}

	defer resp.Body.Close()
	w.WriteHeader(http.StatusOK)
	io.Copy(w, resp.Body)
}

func startHTTPServer() *http.Server {
	// before starting the server determine the port and timeouts
	// for the requests to be proxied using environment variables.
	// The server timeout env var needs to be a string that can be
	// parsed by time.ParseDuration(), so something in the form of
	// 1s, 100ms etc

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = ServerPort
	}
	log.Printf("Server port = %s", port)

	timeout := os.Getenv("TIMEOUT")
	if len(timeout) != 0 {
		tmpTimeout, err := time.ParseDuration(timeout)
		if err == nil {
			ServerTimeout = tmpTimeout
		}
	}
	log.Printf("Server timeout = %s", ServerTimeout)

	// kind of forced to use gorilla mux to avoid the path
	// cleaning done by the standard library http mux which
	// prevents us from specifying URLs with // and results
	// in a 301 redirect
	router := mux.NewRouter()
	router = router.SkipClean(true)
	router.HandleFunc("/proxy/{url:.*}", proxyServerHTTPHandler).Methods("GET")
	router.HandleFunc("/proxy/{url:.*}", proxyServerHTTPHandler).Methods("POST")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("Starting HTTP server")

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	return srv
}

func main() {
	// register the program to wait for SIGINT or SIGTERM
	// after it has started so we can let the server run
	// in the background and wait for the user to terminate
	// the program
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	srv := startHTTPServer()
	<-sigs

	log.Printf("Stopping HTTP Server")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
}
