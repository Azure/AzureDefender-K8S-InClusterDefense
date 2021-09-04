package wrappers

import (
	"crypto/tls"
	"github.com/go-redis/redis"
	"time"
)

// IRedisBaseClientWrapper is a wrapper interface for base client of redis - "github.com/go-redis/redis"
type IRedisBaseClientWrapper interface {
	// Set Redis `SET key value [expiration]` command.
	// Use expiration for `SETEX`-like behavior.
	// Zero expiration means the key has no expiration time.
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd

	// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
	Get(key string) *redis.StringCmd
}

// RedisBaseCacheClientConfiguration redis cache client configuration
type RedisBaseCacheClientConfiguration struct {
	// Addr host:port address.
	Addr string
	// Password Optional password. Must match the password specified in the
	// requirement pass server configuration option.
	Password string
	// Db Database to be selected after connecting to the server.
	Db int
	// MaxRetries Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// MinRetryBackoff Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}
