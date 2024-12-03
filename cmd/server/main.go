package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/theoptz/pow-tcp/internal/pow/hashcash"
	"github.com/theoptz/pow-tcp/internal/server/config"
	"github.com/theoptz/pow-tcp/internal/server/quotes/inmemory"
	"github.com/theoptz/pow-tcp/internal/server/server"
)

const (
	gracefulShutdownTimeout = 10 * time.Second
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load env config")
	}

	log.Debug().Any("config", cfg).Msg("Parsed config")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pw := hashcash.New()

	srv := server.NewServer(*cfg, pw, inmemory.New(), log.With().Str("pkg", "server").Logger())
	if err = srv.Listen(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}

	var pprofServer *http.Server
	if cfg.Pprof.Enabled {
		go func() {
			pprofServer = &http.Server{
				Addr: fmt.Sprintf("%s:%d", cfg.Pprof.Host, cfg.Pprof.Port),
			}
			log.Debug().Any("endpoint", pprofServer.Addr).Msg("Starting pprof server")
			if httpListenErr := pprofServer.ListenAndServe(); httpListenErr != nil && !errors.Is(httpListenErr, http.ErrServerClosed) {
				log.Error().Err(httpListenErr).Msg("failed to start pprof server")
			}
		}()
	}

	<-ctx.Done()

	stopTcpServer(srv)
	stopPprofServer(pprofServer)
}

func stopTcpServer(srv *server.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}

func stopPprofServer(pprofServer *http.Server) {
	if pprofServer == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	if err := pprofServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}
