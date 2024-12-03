package config

import (
	"time"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Clients        int `long:"clients" env:"CLIENTS" description:"Total number of client connections" default:"100"`
	MaxConcurrency int `long:"max-concurrency" env:"MAX_CONCURRENCY" description:"Maximum number of concurrent connections allowed" default:"10"`
	ClientConfig   ClientConfig
}

type ClientConfig struct {
	Host         string        `long:"host" env:"HOST" description:"Hostname or IP address of the server" default:"localhost"`
	Port         int           `long:"port" env:"PORT" description:"Port number for the server connection" default:"10001"`
	DialTimeout  time.Duration `long:"dial-timeout" env:"DIAL_TIMEOUT" description:"Timeout duration for establishing a connection with the server" default:"1s"`
	ReadTimeout  time.Duration `long:"read-timeout" env:"READ_TIMEOUT" description:"Timeout duration for reading from the server" default:"1s"`
	WriteTimeout time.Duration `long:"write-timeout" env:"WRITE_TIMEOUT" description:"Timeout duration for writing to the server" default:"1s"`
}

func FromEnv() (*Config, error) {
	var cfg Config

	parser := flags.NewParser(&cfg, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
