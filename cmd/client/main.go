package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theoptz/pow-tcp/internal/client/client"
	"github.com/theoptz/pow-tcp/internal/client/config"
	"github.com/theoptz/pow-tcp/internal/pow"
	"github.com/theoptz/pow-tcp/internal/pow/hashcash"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load env config")
	}

	log.Debug().Any("config", cfg).Msg("Parsed config")

	solver := hashcash.New()
	logger := log.With().Str("pkg", "client").Logger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	jobChan := make(chan struct{}, cfg.MaxConcurrency)

	for i := 0; i < cfg.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Simulate distributed load over time to mimic real-world behavior
			sleepMs := rand.IntN(50)
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)

			for range jobChan {
				if runErr := runClient(ctx, cfg.ClientConfig, solver, logger); runErr != nil {
					log.Error().Err(runErr).Msg("failed to communicate with the server")
				}
			}
		}()
	}

	for i := 0; i < cfg.Clients; i++ {
		jobChan <- struct{}{}
	}

	close(jobChan)

	wg.Wait()
}

func runClient(ctx context.Context, cfg config.ClientConfig, solver pow.Solver, logger zerolog.Logger) error {
	cl, err := client.New(cfg, solver, logger)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := cl.Close(); closeErr != nil {
			logger.Error().Err(closeErr).Msg("failed to close client")
		}
	}()

	if err = cl.Do(ctx); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}
