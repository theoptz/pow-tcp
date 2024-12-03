package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/theoptz/pow-tcp/internal/pow"
	"github.com/theoptz/pow-tcp/internal/server/config"
	"github.com/theoptz/pow-tcp/internal/server/quotes"
)

const (
	defaultTimeout = time.Second
)

type Server struct {
	endpoint      string
	listener      net.Listener
	wg            sync.WaitGroup
	once          sync.Once
	done          chan struct{}
	concurrency   chan struct{}
	acceptTimeout time.Duration
	pw            pow.ProofOfWork
	powCfg        config.PowConfig
	quote         quotes.Quote
	logger        zerolog.Logger
}

func NewServer(cfg config.Config, pw pow.ProofOfWork, quote quotes.Quote, logger zerolog.Logger) *Server {
	return &Server{
		endpoint:      fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		done:          make(chan struct{}),
		concurrency:   make(chan struct{}, cfg.MaxConcurrency),
		acceptTimeout: cfg.AcceptTimeout,
		pw:            pw,
		powCfg:        cfg.Pow,
		quote:         quote,
		logger:        logger,
	}
}

func (s *Server) Listen() error {
	var err error
	if s.listener, err = net.Listen("tcp", s.endpoint); err != nil {
		return fmt.Errorf("failed to listen %s: %w", s.endpoint, err)
	}

	s.logger.Debug().Str("endpoint", s.endpoint).Msg("Listening")

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Debug().Msg("Shutting down")

	var err error
	s.once.Do(func() {
		close(s.done)

		if s.listener != nil {
			if err = s.listener.Close(); err != nil {
				s.logger.Error().Err(err).Msg("failed to close listener")
			}
		}
	})

	shutdownCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(shutdownCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-shutdownCh:
	}

	return err
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return
		default:
			break
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			s.logger.Error().Err(err).Msg("failed to accept connection")
			break
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	var err error
	defer func() {
		s.wg.Done()

		s.logger.Debug().Err(err).Msg("Closing connection")
		if err := conn.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close connection")
		}
	}()

	t := time.NewTimer(s.acceptTimeout)
	select {
	case <-s.done:
		t.Stop()
		return
	case s.concurrency <- struct{}{}:
		t.Stop()
		defer func() {
			<-s.concurrency
		}()
	case <-t.C:
		s.logger.Error().Msg("max concurrency limit reached")
		return
	}

	s.logger.Debug().Msg("New connection accepted. Initialize challenge")

	if err = s.checkPow(conn); err != nil {
		s.logger.Error().Err(err).Msg("failed to proof connection")
		return
	}

	s.logger.Debug().Msg("Solution accepted")

	if err = s.processConnection(conn); err != nil {
		s.logger.Error().Err(err).Msg("failed to process connection")
		return
	}
}

func (s *Server) checkPow(conn net.Conn) error {
	challenge, err := s.pw.GenerateChallenge(s.powCfg.Difficulty)
	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}

	if err = conn.SetWriteDeadline(time.Now().Add(s.powCfg.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err = conn.Write(challenge[:]); err != nil {
		return fmt.Errorf("failed to write challenge: %w", err)
	}

	if err = conn.SetReadDeadline(time.Now().Add(s.powCfg.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	solution := make([]byte, 8)
	if _, err = conn.Read(solution); err != nil {
		return fmt.Errorf("failed to read buffer: %w", err)
	}

	if ok := s.pw.VerifySolution(challenge, solution); !ok {
		return fmt.Errorf("invalid solution")
	}

	return nil
}

func (s *Server) processConnection(conn net.Conn) error {
	// Use 1s as the default timeout here and within this function for simplicity
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	quote, err := s.quote.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	// Ensure connection deadlines are updated accordingly
	if err = conn.SetWriteDeadline(time.Now().Add(defaultTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	payload := make([]byte, len(quote)+1)
	copy(payload, quote)
	payload[len(quote)] = '\n'

	if _, err = conn.Write(payload); err != nil {
		return fmt.Errorf("failed to write quote: %w", err)
	}

	return nil
}
