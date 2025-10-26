# Go Redis Kit

A simple, production-ready Redis client wrapper for Go with easy configuration and connection pooling.

```go
import "github.com/alinemone/go-redis-kit"

client, _ := rediskit.New(nil)
client.Set(ctx, "key", "value", 0)
```

## Features

- ✅ Simple wrapper around [go-redis/redis](https://github.com/redis/go-redis)
- ✅ Easy configuration with sensible defaults
- ✅ Connection pooling with configurable settings
- ✅ Built-in health check functionality
- ✅ Configuration validation
- ✅ Thread-safe operations
- ✅ Access to all go-redis methods via embedding

## Installation

```bash
go get github.com/alinemone/go-redis-kit
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/alinemone/go-redis-kit"
)

func main() {
    // Create client with default configuration (localhost:6379)
    client, err := rediskit.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Use any go-redis method directly
    err = client.Set(ctx, "key", "value", 0).Err()
    if err != nil {
        log.Fatal(err)
    }

    val, err := client.Get(ctx, "key").Result()
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Value:", val)
}
```

### Custom Configuration

```go
cfg := &rediskit.Config{
    Host:                 "localhost",
    Port:                 "6379",
    Password:             "your-password",
    DB:                   0,
    PoolSize:             20,
    MinIdleConns:         5,
    ConnMaxIdleTime:      10 * time.Minute,
    ConnMaxLifetime:      30 * time.Minute,
    SocketTimeout:        5 * time.Second,
    SocketConnectTimeout: 3 * time.Second,
    DefaultTimeout:       5 * time.Second,
    MaxRetries:           3,
    MinRetryBackoff:      100 * time.Millisecond,
    MaxRetryBackoff:      1 * time.Second,
}

client, err := rediskit.New(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## API Reference

### Creating a Client

#### `New(cfg *Config) (*Client, error)`

Creates a new Redis client with the given configuration. If `cfg` is `nil`, uses default configuration.

```go
client, err := rediskit.New(nil) // Use defaults
// OR
cfg := rediskit.DefaultConfig()
cfg.Host = "redis-server"
cfg.PoolSize = 50
client, err := rediskit.New(cfg)
```

#### `DefaultConfig() *Config`

Returns a configuration with sensible defaults:
- Host: `localhost`
- Port: `6379`
- PoolSize: `10`
- MinIdleConns: `2`
- DefaultTimeout: `5s`
- SocketTimeout: `5s`
- SocketConnectTimeout: `3s`
- MaxRetries: `5`
- ConnMaxIdleTime: `5m`
- ConnMaxLifetime: `30m`

### Client Methods

#### `HealthCheck() error`

Performs a health check by pinging the Redis server.

```go
if err := client.HealthCheck(); err != nil {
    log.Fatal("Redis is not healthy:", err)
}
```

#### `GetConfig() *Config`

Returns the client configuration.

```go
cfg := client.GetConfig()
fmt.Println("Pool size:", cfg.PoolSize)
```

### Using Redis Commands

Since `Client` embeds `*redis.Client`, you have access to **all go-redis methods** directly:

```go
// String operations
client.Set(ctx, "key", "value", time.Hour).Err()
client.Get(ctx, "key").Result()
client.Del(ctx, "key1", "key2").Err()
client.Incr(ctx, "counter").Result()

// Hash operations
client.HSet(ctx, "hash", "field", "value").Err()
client.HGet(ctx, "hash", "field").Result()
client.HGetAll(ctx, "hash").Result()

// List operations
client.LPush(ctx, "list", "item").Err()
client.LRange(ctx, "list", 0, -1).Result()

// Set operations
client.SAdd(ctx, "set", "member").Err()
client.SMembers(ctx, "set").Result()

// And many more... see go-redis documentation
```

### Error Handling

```go
import "github.com/redis/go-redis/v9"

val, err := client.Get(ctx, "key").Result()
if err != nil {
    if errors.Is(err, redis.Nil) {
        // Key does not exist
        log.Println("Key not found")
    } else {
        // Other error
        log.Fatal("Redis error:", err)
    }
}
```

Package-specific errors:

```go
var (
    ErrNilClient     = errors.New("redis client is nil")
    ErrInvalidConfig = errors.New("invalid redis configuration")
)
```

## Advanced Usage

### Connection Pooling

The client automatically manages a connection pool. Configure it based on your needs:

```go
cfg := rediskit.DefaultConfig()
cfg.PoolSize = 50              // Maximum number of connections
cfg.MinIdleConns = 10          // Minimum idle connections
cfg.ConnMaxIdleTime = 5 * time.Minute
cfg.ConnMaxLifetime = 30 * time.Minute
```

### Timeouts and Retries

```go
cfg := rediskit.DefaultConfig()
cfg.SocketTimeout = 10 * time.Second         // Read/Write timeout
cfg.SocketConnectTimeout = 5 * time.Second   // Connection timeout
cfg.DefaultTimeout = 5 * time.Second         // Default timeout for HealthCheck
cfg.MaxRetries = 3                           // Maximum retry attempts
cfg.MinRetryBackoff = 100 * time.Millisecond
cfg.MaxRetryBackoff = 1 * time.Second
```

### Direct Access to go-redis Client

The underlying `*redis.Client` is embedded, so you have full access:

```go
client, err := rediskit.New(cfg)
if err != nil {
    log.Fatal(err)
}

// Use any redis.Client method
result := client.Do(ctx, "CUSTOM", "COMMAND")
pipeline := client.Pipeline()
// etc...
```

## Best Practices

1. **Always use context**: Pass a proper context for cancellation and timeout control
2. **Reuse clients**: Create one client and reuse it throughout your application
3. **Handle redis.Nil**: Always check for `redis.Nil` error when getting keys
4. **Configure pool size**: Set pool size based on your expected concurrent operations
5. **Use health checks**: Implement health checks in your application monitoring
6. **Close connections**: Always defer `client.Close()` when done

## Example: Web Application

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/alinemone/go-redis-kit"
)

var redisClient *rediskit.Client

func init() {
    var err error
    redisClient, err = rediskit.New(nil)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }

    // Verify connection
    if err := redisClient.HealthCheck(); err != nil {
        log.Fatal("Redis health check failed:", err)
    }
}

func main() {
    http.HandleFunc("/set", setHandler)
    http.HandleFunc("/get", getHandler)
    http.HandleFunc("/health", healthHandler)

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func setHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    key := r.URL.Query().Get("key")
    value := r.URL.Query().Get("value")

    if err := redisClient.Set(ctx, key, value, time.Hour).Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write([]byte("OK"))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    key := r.URL.Query().Get("key")

    value, err := redisClient.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            http.Error(w, "Key not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Write([]byte(value))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    if err := redisClient.HealthCheck(); err != nil {
        http.Error(w, "Redis unhealthy", http.StatusServiceUnavailable)
        return
    }

    w.Write([]byte("OK"))
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
