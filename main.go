package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelError,
	})))

	var yamlPath string
	flag.StringVar(&yamlPath, "configpath", defaultYamlConfigPath, fmt.Sprintf("path to config yaml file, default: %s", defaultYamlConfigPath))
	flag.Parse()

	cfg, err := loadConfig(yamlPath)
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	vmAPI := NewAPI()

	// routes
	mux := http.NewServeMux()
	mux.HandleFunc("/addvm", vmAPI.AddVMHandler)
	mux.HandleFunc("/insert", vmAPI.InsertCoinHandler)
	mux.HandleFunc("/select", vmAPI.SelectProductHandler)
	mux.HandleFunc("/abort", vmAPI.AbortOrderHandler)

	// statemachine routes
	mux.HandleFunc("/sm/insert", vmAPI.TransitionHandler)
	mux.HandleFunc("/sm/select", vmAPI.TransitionHandler)

	// serve
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	serverCtx, cancel := context.WithCancelCause(context.Background())

	go func() {
		log.Printf("listening on %s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %v\n", err)
		}

		cancel(err)
	}()

	doneCh := make(chan struct{})

	// graceful shutdown
	go func() {
		<-serverCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %v\n", err)
		}

		close(doneCh)
	}()

	<-doneCh
}
