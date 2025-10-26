package rediskit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNilClient     = errors.New("redis client is nil")
	ErrInvalidConfig = errors.New("invalid redis configuration")
)

// Config holds Redis client configuration
type Config struct {
	Host                 string
	Port                 string
	Password             string
	DB                   int
	SocketKeepalive      bool
	HealthCheckInterval  time.Duration
	SocketTimeout        time.Duration
	SocketConnectTimeout time.Duration
	MaxRetries           int
	MinRetryBackoff      time.Duration
	MaxRetryBackoff      time.Duration
	PoolSize             int
	MinIdleConns         int
	ConnMaxIdleTime      time.Duration
	ConnMaxLifetime      time.Duration
	DefaultTimeout       time.Duration // Default timeout for operations
}

func DefaultConfig() *Config {
	return &Config{
		Host:                 "localhost",
		Port:                 "6379",
		SocketKeepalive:      true,
		HealthCheckInterval:  5 * time.Second,
		SocketTimeout:        5 * time.Second,
		SocketConnectTimeout: 3 * time.Second,
		MaxRetries:           5,
		MinRetryBackoff:      100 * time.Millisecond,
		MaxRetryBackoff:      100 * time.Millisecond,
		PoolSize:             10,
		MinIdleConns:         2,
		ConnMaxIdleTime:      5 * time.Minute,
		ConnMaxLifetime:      30 * time.Minute,
		DefaultTimeout:       5 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidConfig)
	}
	if c.Port == "" {
		return fmt.Errorf("%w: port is required", ErrInvalidConfig)
	}
	if c.PoolSize <= 0 {
		return fmt.Errorf("%w: pool size must be greater than 0", ErrInvalidConfig)
	}
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("%w: default timeout must be greater than 0", ErrInvalidConfig)
	}
	return nil
}

// Client wraps redis.Client with additional functionality
type Client struct {
	*redis.Client
	config *Config
}

// New creates a new Redis client with the given configuration
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Host + ":" + cfg.Port,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		DialTimeout:     cfg.SocketConnectTimeout,
		ReadTimeout:     cfg.SocketTimeout,
		WriteTimeout:    cfg.SocketTimeout,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	return &Client{
		Client: rdb,
		config: cfg,
	}, nil
}

// HealthCheck performs a health check on the Redis connection
func (c *Client) HealthCheck() error {
	if c.Client == nil {
		return ErrNilClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.config.DefaultTimeout)
	defer cancel()
	return c.Client.Ping(ctx).Err()
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.config
}
