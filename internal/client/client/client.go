package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/theoptz/pow-tcp/internal/client/config"
	"github.com/theoptz/pow-tcp/internal/pow"
)

func New(cfg config.ClientConfig, solver pow.Solver, logger zerolog.Logger) (*Client, error) {
	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := net.DialTimeout("tcp", endpoint, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to endpoint %s: %w", endpoint, err)
	}

	return &Client{
		endpoint: endpoint,
		cfg:      cfg,
		conn:     conn,
		solver:   solver,
		logger:   logger,
	}, nil
}

type Client struct {
	endpoint string
	cfg      config.ClientConfig
	conn     net.Conn
	solver   pow.Solver
	logger   zerolog.Logger
}

func (c *Client) Do(ctx context.Context) error {
	var err error
	if err = c.connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err = c.conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	quote, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("solution declined")
		}

		return fmt.Errorf("failed to read quote: %w", err)
	}

	c.logger.Debug().Str("quote", strings.TrimRight(quote, "\n")).Msg("Solution accepted")

	return nil
}

func (c *Client) Close() error {
	c.logger.Debug().Msg("Closing connection")
	return c.conn.Close()
}

func (c *Client) connect(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var err error
	if err = c.conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	c.logger.Debug().Msg("Reading challenge")

	var challenge pow.Challenge
	if _, err = c.conn.Read(challenge[:]); err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}

	c.logger.Debug().Uint8("difficulty", challenge[0]).Msg("Challenge received")

	ctx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	solution, err := c.solver.Solve(ctx, challenge)
	if err != nil {
		return fmt.Errorf("failed to solve challenge: %w", err)
	}

	// Simulate failure of proof of work every tenth request by setting a predefined solution
	if rand.IntN(10) == 0 {
		solution = []byte("12345678")
	}

	c.logger.Debug().Msg("Sending solution")
	if err = c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err = c.conn.Write(solution); err != nil {
		return fmt.Errorf("failed to write solution: %w", err)
	}

	return nil
}
