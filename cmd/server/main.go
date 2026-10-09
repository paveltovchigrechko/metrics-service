package main

import (
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/middleware"
	"github.com/paveltovchigrechko/metrics-service/internal/server"
)

// pprofAddrEnv is the environment variable with the address of the pprof HTTP server, e.g. "localhost:6060".
// If it is empty, profiling is disabled.
const pprofAddrEnv = "PPROF_ADDR"

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.SetServerConfig(os.Args[1:])
	if err != nil {
		return err
	}

	l, err := logger.New()
	if err != nil {
		return err
	}
	defer l.Sync() // Flush at the end

	startPprof(os.Getenv(pprofAddrEnv))

	middlewares := []func(http.Handler) http.Handler{
		logger.LoggerMiddleware(l),
	}

	if cfg.Key != "" {
		middlewares = append(middlewares,
			middleware.VerifyHashMiddleware(cfg.Key),
			middleware.SignResponseMiddleware(cfg.Key),
		)
	}

	middlewares = append(middlewares, middleware.GZIPMiddleware)

	serv, err := server.New(cfg, middlewares...)

	if err != nil {
		return err
	}
	defer serv.StopDB()
	defer serv.StopAudit() // Runs before StopDB: flush audit events after the server stopped.

	err = serv.Run()
	if err != nil {
		return err
	}

	return nil
}

// startPprof serves the pprof endpoints (/debug/pprof/) on a separate address in the background.
// It uses its own mux instead of http.DefaultServeMux, so the profiling handlers are never exposed on the public router.
// If addr is empty, nothing is started.
func startPprof(addr string) {
	if addr == "" {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("pprof is listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("pprof server error: %v", err)
		}
	}()
}
