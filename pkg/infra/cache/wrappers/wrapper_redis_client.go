package wrappers

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// redis.Client implements IRedisBaseClientWrapper interface.
var _ IRedisBaseClientWrapper = (*redis.Client)(nil)

// IRedisBaseClientWrapper is a wrapper interface for base client of redis - "github.com/go-redis/redis"
type IRedisBaseClientWrapper interface {
	// Set Redis `SET key value [expiration]` command.
	// Use expiration for `SETEX`-like behavior.
	// Zero expiration means the key has no expiration time.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd

	// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
	Get(ctx context.Context, key string) *redis.StringCmd

	// Ping Redis Ping command. it is used to test if a connection is still alive, or to measure latency.
	// returns redis.Nil error if successfully received a pong from the server.
	Ping(ctx context.Context) *redis.StatusCmd
}

// RedisCacheClientConfiguration redis cache client configuration
type RedisCacheClientConfiguration struct {
	// Address host:port address.
	Address string
	// Host address
	Host string
	// Password Optional password. Must match the password specified in the
	// requirement pass server configuration option.
	PasswordPath string
	// TlsCrt is the path to the tls.crt file
	TlsCrtPath string
	// TlsKey is the path to the tls.key file
	TlsKeyPath string
	// CaCert is the path to the ca.cert file
	CaCertPath string
	// Table is Database to be selected after connecting to the server.
	Table int
	// MaxRetries Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// MinRetryBackoff Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
}
