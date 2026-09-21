package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"hotelbooking/internal/api"
	"hotelbooking/internal/booking"
	"hotelbooking/internal/config"
	"hotelbooking/internal/hotel"
	"hotelbooking/internal/store"
)

// shutdownTimeout is how long in-flight requests get to finish on shutdown
const shutdownTimeout = 10 * time.Second

func main() {

	cfg := config.Config{}

	err := envconfig.Process("", &cfg)
	if err != nil {
		fatal("could not load the configuration", "err", err)
	}

	logHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: cfg.LogLevel,
		})
	slog.SetDefault(slog.New(logHandler))

	st, err := store.NewStore(cfg.Postgres)
	if err != nil {
		fatal("could not open the database", "err", err)
	}

	service := &booking.Service{DB: st.DB}

	server := &api.Server{
		Store:           st,
		Hotels:          &hotel.Service{DB: st.DB},
		BookingsService: service,
	}

	e := server.Routes()

	go func() {
		slog.Info("api listening", "port", cfg.Port)

		err := e.Start(":" + cfg.Port)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal("could not start the api", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = e.Shutdown(ctx)
	if err != nil {
		slog.Error("could not shut down cleanly :(", "err", err)
	}

	service.Wait()

	slog.Info("stopped")
}

func fatal(msg string, args ...any) {

	slog.Error(msg, args...)
	os.Exit(1)
}
