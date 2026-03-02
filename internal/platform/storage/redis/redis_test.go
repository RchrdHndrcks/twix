package redis_test

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// setupRedis starts a Redis container and returns a connected client.
// It skips the test if Docker is not available.
func setupRedis(t *testing.T) *goredis.Client {
	t.Helper()
	ctx := context.Background()

	container, err := tryStartRedis(t, ctx)
	if err != nil {
		t.Skipf("skipping: Docker not available (%v)", err)
	}

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}

	opts, err := goredis.ParseURL(connStr)
	if err != nil {
		t.Fatalf("failed to parse redis URL: %v", err)
	}

	client := goredis.NewClient(opts)
	t.Cleanup(func() {
		client.FlushAll(ctx)
		client.Close()
		_ = container.Terminate(ctx)
	})

	return client
}

// tryStartRedis attempts to start a Redis container, recovering from panics
// caused by Docker not being available (testcontainers panics internally on first call).
func tryStartRedis(t *testing.T, ctx context.Context) (container *tcredis.RedisContainer, err error) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("skipping: Docker not available (%v)", r)
		}
	}()

	container, err = tcredis.Run(ctx, "redis:7-alpine")
	return container, err
}
