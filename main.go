package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"vendingmachine/internal/storage"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelError,
	})))

	var yamlPath string
	defaultYamlConfigPath := "./config.yaml"
	flag.StringVar(&yamlPath, "configpath", defaultYamlConfigPath,
		fmt.Sprintf("path to config yaml file, default: %s", defaultYamlConfigPath))
	flag.Parse()

	cfg, err := loadConfig(yamlPath)
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	vmStorage := storage.NewInMemoryVMStorage()
	smStorage := storage.NewInMemorySMStorage()

	handler := NewHandler(vmStorage, smStorage)

	// routes
	mux := http.NewServeMux()
	mux.HandleFunc("/addvm", handler.AddVMHandler)
	mux.HandleFunc("/insert", handler.InsertCoinHandler)
	mux.HandleFunc("/select", handler.SelectProductHandler)
	mux.HandleFunc("/abort", handler.AbortOrderHandler)

	// statemachine routes
	mux.HandleFunc("/sm/insert", handler.TransitionHandler)
	mux.HandleFunc("/sm/select", handler.TransitionHandler)

	// serve
	srv := http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeoutSeconds) * time.Second,
	}

	serverCtx, cancel := context.WithCancelCause(context.Background())

	go func() {
		log.Printf("listening on %s\n", srv.Addr)
		err := srv.ListenAndServe() //nolint: govet // shadowing is not a problem here
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %v\n", err)
		}

		cancel(err)
	}()

	doneCh := make(chan struct{})

	// graceful shutdown
	go func() {
		<-serverCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(),
			time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil { //nolint: govet // shadowing is not a problem here
			fmt.Fprintf(os.Stderr, "error shutting down http server: %v\n", err)
		}

		close(doneCh)
	}()

	<-doneCh
}
