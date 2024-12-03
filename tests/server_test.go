package tests

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/theoptz/pow-tcp/internal/client/client"
	"github.com/theoptz/pow-tcp/internal/client/config"
	"github.com/theoptz/pow-tcp/internal/pow"
	"github.com/theoptz/pow-tcp/internal/pow/hashcash"
)

const (
	defaultTimeout = time.Second * 5
)

func TestServer(t *testing.T) {
	suite.Run(t, &ServerSuite{})
}

type ServerSuite struct {
	suite.Suite

	cfg    config.ClientConfig
	logger zerolog.Logger
}

func (s *ServerSuite) SetupSuite() {
	s.cfg = getDefaultConfig()
	s.logger = zerolog.Nop()
}

func (s *ServerSuite) TestSuccessPow() {
	solver := hashcash.New()
	cl, err := client.New(s.cfg, solver, s.logger)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	assert.NoError(s.T(), cl.Do(ctx))
	assert.NoError(s.T(), cl.Close())
}

func (s *ServerSuite) TestFailedPow() {
	solver := new(nopSolver)
	cl, err := client.New(s.cfg, solver, s.logger)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	assert.Error(s.T(), cl.Do(ctx))
	assert.NoError(s.T(), cl.Close())
}

func getDefaultConfig() config.ClientConfig {
	return config.ClientConfig{
		Host:         getEnvOrDefault("HOST", "localhost"),
		Port:         toInt(getEnvOrDefault("PORT", "10001")),
		DialTimeout:  toDuration(getEnvOrDefault("DIAL_TIMEOUT", "1s")),
		ReadTimeout:  toDuration(getEnvOrDefault("READ_TIMEOUT", "1s")),
		WriteTimeout: toDuration(getEnvOrDefault("WRITE_TIMEOUT", "1s")),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	res := os.Getenv(key)
	if res == "" {
		return defaultValue
	}

	return res
}

func toDuration(val string) time.Duration {
	res, err := time.ParseDuration(val)
	if err != nil {
		panic(err)
	}

	return res
}

func toInt(val string) int {
	res, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	return res
}

type nopSolver struct{}

func (n *nopSolver) Solve(ctx context.Context, challenge pow.Challenge) ([]byte, error) {
	return []byte("12345678"), nil
}
