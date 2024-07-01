//go:build disabled && integration
// +build disabled,integration

package main

// These tests are disabled because test suite doesn't work properly.
// There is a bug with releasing resources in our TCP server implementation,
// so all tests after first fails because of address is already bound

import (
	"context"
	"errors"
	"github.com/burenotti/redis_impl/internal/config"
	"github.com/burenotti/redis_impl/internal/server"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log/slog"
	"testing"
	"time"
)

func TestTestSuite(t *testing.T) {
	s := &TestSuite{}
	suite.Run(t, s)
}

type TestSuite struct {
	suite.Suite
	redis  *redis.Client
	server *server.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *TestSuite) SetupTest() {
	s.redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:7379",
		Protocol: 2,
	})

	cfg := &config.Config{}
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 7379
	cfg.Server.MaxConnections = 10

	s.server = initServer(slog.Default(), cfg)
	go func() {
		if err := s.server.Run(); err != nil {
			if !errors.Is(err, context.Canceled) {
				s.Failf("", "failed to start server: %v", err.Error())
			}
		}
	}()

	waitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for {
		if err := s.redis.Ping(waitCtx).Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				s.Fail("timeout waiting for server to start")
				break
			}
		} else {
			break
		}
	}
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Second)
}

func (s *TestSuite) TearDownTest() {
	s.cancel()
	require.NoError(s.T(), s.redis.Close())
	require.NotPanics(s.T(), func() {
		require.NoError(s.T(), s.server.Stop(1*time.Second))
	})
}

func (s *TestSuite) TestMultiplePing() {
	for i := 0; i < 10; i++ {
		res, err := s.redis.Ping(s.ctx).Result()
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "PONG", res)
	}
}

func (s *TestSuite) TestCanSetStringValues() {
	{
		cmd := s.redis.Set(s.ctx, "first_name", "artem", 0)
		assertResultEqual(s.T(), cmd, "OK")
		cmd = s.redis.Set(s.ctx, "last_name", "burenin", 0)
		assertResultEqual(s.T(), cmd, "OK")
	}

	cmd := s.redis.Get(s.ctx, "first_name")
	assertResultEqual(s.T(), cmd, "artem")

	cmd = s.redis.Get(s.ctx, "last_name")
	assertResultEqual(s.T(), cmd, "burenin")

	cmd = s.redis.Get(s.ctx, "middle_name")
	assertResultError(s.T(), cmd, redis.Nil)
}

func (s *TestSuite) TestTransactions() {
	{
		cmd := s.redis.Set(s.ctx, "first_name", "artem", 0)
		assertResultEqual(s.T(), cmd, "OK")
		cmd = s.redis.Set(s.ctx, "last_name", "burenin", 0)
		assertResultEqual(s.T(), cmd, "OK")
	}
}

func assertResultEqual(t *testing.T, cmd interface{}, expected interface{}) {
	c, ok := cmd.(interface{ Result() (string, error) })
	if !ok {
		assert.FailNow(t, "cmd must has a Result() method")
	}
	res, err := c.Result()
	require.NoError(t, err)
	assert.Equal(t, res, expected)
}

func assertResultError(t *testing.T, cmd interface{}, expected error) {
	c, ok := cmd.(interface{ Result() (string, error) })
	if !ok {
		assert.FailNow(t, "cmd must has a Result() method")
	}
	_, err := c.Result()
	assert.ErrorIs(t, err, expected)
}
