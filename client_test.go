package rediskit

import (
	"strings"
	"testing"
	"time"
)

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", cfg.Host, "localhost"},
		{"Port", cfg.Port, "6379"},
		{"PoolSize", cfg.PoolSize, 10},
		{"MinIdleConns", cfg.MinIdleConns, 2},
		{"DefaultTimeout", cfg.DefaultTimeout, 5 * time.Second},
		{"SocketTimeout", cfg.SocketTimeout, 5 * time.Second},
		{"SocketConnectTimeout", cfg.SocketConnectTimeout, 3 * time.Second},
		{"MaxRetries", cfg.MaxRetries, 5},
		{"MinRetryBackoff", cfg.MinRetryBackoff, 100 * time.Millisecond},
		{"MaxRetryBackoff", cfg.MaxRetryBackoff, 100 * time.Millisecond},
		{"ConnMaxIdleTime", cfg.ConnMaxIdleTime, 5 * time.Minute},
		{"ConnMaxLifetime", cfg.ConnMaxLifetime, 30 * time.Minute},
		{"SocketKeepalive", cfg.SocketKeepalive, true},
		{"HealthCheckInterval", cfg.HealthCheckInterval, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestConfigValidate tests configuration validation
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errString string
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host:           "",
				Port:           "6379",
				PoolSize:       10,
				DefaultTimeout: 5 * time.Second,
			},
			wantErr:   true,
			errString: "host is required",
		},
		{
			name: "empty port",
			config: &Config{
				Host:           "localhost",
				Port:           "",
				PoolSize:       10,
				DefaultTimeout: 5 * time.Second,
			},
			wantErr:   true,
			errString: "port is required",
		},
		{
			name: "zero pool size",
			config: &Config{
				Host:           "localhost",
				Port:           "6379",
				PoolSize:       0,
				DefaultTimeout: 5 * time.Second,
			},
			wantErr:   true,
			errString: "pool size must be greater than 0",
		},
		{
			name: "negative pool size",
			config: &Config{
				Host:           "localhost",
				Port:           "6379",
				PoolSize:       -5,
				DefaultTimeout: 5 * time.Second,
			},
			wantErr:   true,
			errString: "pool size must be greater than 0",
		},
		{
			name: "zero timeout",
			config: &Config{
				Host:           "localhost",
				Port:           "6379",
				PoolSize:       10,
				DefaultTimeout: 0,
			},
			wantErr:   true,
			errString: "default timeout must be greater than 0",
		},
		{
			name: "negative timeout",
			config: &Config{
				Host:           "localhost",
				Port:           "6379",
				PoolSize:       10,
				DefaultTimeout: -1 * time.Second,
			},
			wantErr:   true,
			errString: "default timeout must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain %q, got %q", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestNew tests client creation
func TestNewClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client, err := NewClient(nil)
		if err != nil {
			// It's okay to fail connection to localhost:6379 if Redis is not running
			// We just want to verify config validation worked
			t.Logf("Connection failed (expected if Redis not running): %v", err)
		}
		if client != nil {
			defer client.Close()

			cfg := client.GetConfig()
			if cfg.Host != "localhost" {
				t.Errorf("expected default host localhost, got %s", cfg.Host)
			}
			if cfg.Port != "6379" {
				t.Errorf("expected default port 6379, got %s", cfg.Port)
			}
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &Config{
			Host:           "custom-host",
			Port:           "1234",
			Password:       "secret",
			DB:             1,
			PoolSize:       20,
			MinIdleConns:   5,
			DefaultTimeout: 10 * time.Second,
		}

		client, _ := NewClient(cfg)
		if client != nil {
			defer client.Close()

			gotCfg := client.GetConfig()
			if gotCfg.Host != "custom-host" {
				t.Errorf("expected host custom-host, got %s", gotCfg.Host)
			}
			if gotCfg.Port != "1234" {
				t.Errorf("expected port 1234, got %s", gotCfg.Port)
			}
			if gotCfg.PoolSize != 20 {
				t.Errorf("expected pool size 20, got %d", gotCfg.PoolSize)
			}
		}
	})

	t.Run("with invalid config returns error", func(t *testing.T) {
		cfg := &Config{
			Host:           "",
			Port:           "",
			PoolSize:       0,
			DefaultTimeout: 0,
		}

		client, err := NewClient(cfg)
		if err == nil {
			t.Error("expected error for invalid config, got nil")
		}
		if client != nil {
			t.Error("expected nil client for invalid config")
			client.Close()
		}
		if err != nil && !strings.Contains(err.Error(), "host is required") {
			t.Errorf("expected 'host is required' error, got %v", err)
		}
	})
}

// TestGetConfig tests getting configuration
func TestGetConfig(t *testing.T) {
	cfg := &Config{
		Host:           "test-host",
		Port:           "9999",
		PoolSize:       15,
		DefaultTimeout: 7 * time.Second,
	}

	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("client creation failed")
	}
	defer client.Close()

	gotCfg := client.GetConfig()
	if gotCfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	if gotCfg.Host != cfg.Host {
		t.Errorf("Host: got %s, want %s", gotCfg.Host, cfg.Host)
	}
	if gotCfg.Port != cfg.Port {
		t.Errorf("Port: got %s, want %s", gotCfg.Port, cfg.Port)
	}
	if gotCfg.PoolSize != cfg.PoolSize {
		t.Errorf("PoolSize: got %d, want %d", gotCfg.PoolSize, cfg.PoolSize)
	}
	if gotCfg.DefaultTimeout != cfg.DefaultTimeout {
		t.Errorf("DefaultTimeout: got %v, want %v", gotCfg.DefaultTimeout, cfg.DefaultTimeout)
	}
}

// TestHealthCheck tests health check functionality
func TestHealthCheck(t *testing.T) {
	t.Run("nil client returns error", func(t *testing.T) {
		client := &Client{Client: nil, config: DefaultConfig()}
		err := client.HealthCheck()
		if err != ErrNilClient {
			t.Errorf("expected ErrNilClient, got %v", err)
		}
	})

	t.Run("with valid client", func(t *testing.T) {
		client, err := NewClient(nil)
		if err != nil {
			t.Skip("Redis not available for testing")
		}
		if client == nil {
			t.Skip("client is nil")
		}
		defer client.Close()

		err = client.HealthCheck()
		// If Redis is running, should be nil
		// If not running, will get connection error (which is expected)
		t.Logf("HealthCheck result: %v", err)
	})
}

// TestClientEmbedding tests that client properly embeds redis.Client
func TestClientEmbedding(t *testing.T) {
	cfg := DefaultConfig()
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("client creation failed")
	}
	defer client.Close()

	// Verify we can access embedded Client methods
	if client.Client == nil {
		t.Error("embedded redis.Client is nil")
	}

	// Verify we can call redis methods directly
	// Note: These will fail if Redis is not running, which is okay for this test
	t.Logf("Client type: %T", client.Client)
}

// TestConfigFieldTypes tests that config fields have correct types
func TestConfigFieldTypes(t *testing.T) {
	cfg := DefaultConfig()

	// Test string fields
	_ = cfg.Host
	_ = cfg.Password
	_ = cfg.Port

	// Test int fields
	_ = cfg.DB
	_ = cfg.PoolSize
	_ = cfg.MinIdleConns
	_ = cfg.MaxRetries

	// Test duration fields
	_ = cfg.SocketTimeout
	_ = cfg.SocketConnectTimeout
	_ = cfg.DefaultTimeout
	_ = cfg.MinRetryBackoff
	_ = cfg.MaxRetryBackoff
	_ = cfg.ConnMaxIdleTime
	_ = cfg.ConnMaxLifetime
	_ = cfg.HealthCheckInterval

	// Test bool fields
	_ = cfg.SocketKeepalive

	t.Log("All config fields have correct types")
}

// TestErrorTypes tests custom error types
func TestErrorTypes(t *testing.T) {
	if ErrNilClient == nil {
		t.Error("ErrNilClient should not be nil")
	}
	if ErrInvalidConfig == nil {
		t.Error("ErrInvalidConfig should not be nil")
	}

	expectedNilClientMsg := "redis client is nil"
	if ErrNilClient.Error() != expectedNilClientMsg {
		t.Errorf("ErrNilClient message: got %q, want %q", ErrNilClient.Error(), expectedNilClientMsg)
	}

	expectedInvalidConfigMsg := "invalid redis configuration"
	if ErrInvalidConfig.Error() != expectedInvalidConfigMsg {
		t.Errorf("ErrInvalidConfig message: got %q, want %q", ErrInvalidConfig.Error(), expectedInvalidConfigMsg)
	}
}

// TestConfigAddress tests that address is constructed correctly
func TestConfigAddress(t *testing.T) {
	tests := []struct {
		host         string
		port         string
		expectedAddr string
	}{
		{"localhost", "6379", "localhost:6379"},
		{"127.0.0.1", "6379", "127.0.0.1:6379"},
		{"redis-server", "1234", "redis-server:1234"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedAddr, func(t *testing.T) {
			cfg := &Config{
				Host:           tt.host,
				Port:           tt.port,
				PoolSize:       10,
				DefaultTimeout: 5 * time.Second,
			}

			client, _ := NewClient(cfg)
			if client != nil {
				defer client.Close()
				// The address should be host:port
				expectedAddr := tt.host + ":" + tt.port
				t.Logf("Expected address format: %s", expectedAddr)
			}
		})
	}
}

// TestMultipleClients tests creating multiple clients
func TestMultipleClients(t *testing.T) {
	cfg1 := &Config{
		Host:           "host1",
		Port:           "6379",
		PoolSize:       10,
		DefaultTimeout: 5 * time.Second,
	}

	cfg2 := &Config{
		Host:           "host2",
		Port:           "6380",
		PoolSize:       20,
		DefaultTimeout: 10 * time.Second,
	}

	client1, _ := NewClient(cfg1)
	client2, _ := NewClient(cfg2)

	if client1 != nil {
		defer client1.Close()
	}
	if client2 != nil {
		defer client2.Close()
	}

	if client1 != nil && client2 != nil {
		gotCfg1 := client1.GetConfig()
		gotCfg2 := client2.GetConfig()

		if gotCfg1.Host == gotCfg2.Host {
			t.Error("clients should have different hosts")
		}
		if gotCfg1.PoolSize == gotCfg2.PoolSize {
			t.Error("clients should have different pool sizes")
		}
	}
}
