package config

import (
	"time"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Host           string        `long:"host" env:"HOST" description:"Hostname or IP address to bind the tcp server to" default:"localhost"`
	Port           int           `long:"port" env:"PORT" description:"Port number on which the tcp server will listen" default:"10001"`
	MaxConcurrency int           `long:"max-concurrency" env:"MAX_CONCURRENCY" description:"Maximum number of simultaneous connections the server can handle" default:"1000"`
	AcceptTimeout  time.Duration `long:"accept-timeout" env:"ACCEPT_TIMEOUT" description:"Maximum time to wait for accepting a new connection" default:"1ms"`

	Pow   PowConfig   `group:"pow" namespace:"pow" env-namespace:"POW"`
	Pprof PprofConfig `group:"pprof" namespace:"pprof" env-namespace:"PPROF"`
}

type PowConfig struct {
	Difficulty   uint8         `long:"difficulty" env:"DIFFICULTY" description:"Difficulty level for the Proof of Work challenge (number of leading zero bits)" default:"8"`
	WriteTimeout time.Duration `long:"write-timeout" env:"WRITE_TIMEOUT" description:"Maximum time allowed to send a challenge to the client" default:"250ms"`
	ReadTimeout  time.Duration `long:"read-timeout" env:"READ_TIMEOUT" description:"Maximum time allowed to receive the client's solution for the challenge" default:"500ms"`
}

type PprofConfig struct {
	Enabled bool   `long:"enabled" env:"ENABLED" description:"Flag indicating whether the pprof server is enabled"`
	Host    string `long:"host" env:"HOST" description:"Hostname or IP address to bind the pprof server to" default:"localhost"`
	Port    int    `long:"port" env:"PORT" description:"Port number on which the pprof server will listen" default:"6060"`
}

func FromEnv() (*Config, error) {
	var cfg Config

	parser := flags.NewParser(&cfg, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
