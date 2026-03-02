// Package main is the entry point for the Twix API server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RchrdHndrcks/twix/cmd/api/internal/handlers"
	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/RchrdHndrcks/twix/internal/platform/config"
	"github.com/RchrdHndrcks/twix/internal/platform/storage/memory"
	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
	"github.com/RchrdHndrcks/twix/internal/platform/web"
	"github.com/RchrdHndrcks/twix/internal/timeline"
	"github.com/RchrdHndrcks/twix/internal/tweet"
	"github.com/RchrdHndrcks/twix/internal/user"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		for range 10 {
			// Print the error to stderr and sleep before exiting to
			// give users a chance to see the message.
			fmt.Fprintf(os.Stderr, "error booting application: %v\n", err)
			time.Sleep(1 * time.Second)
		}

		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	var (
		userStore     user.Store
		tweetStore    tweet.Store
		followStore   follow.Store
		timelineStore timeline.Store
	)

	switch cfg.Storage {
	case config.StorageRedis:
		client := redis.NewClient(&redis.Options{
			Addr: cfg.RedisURL,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis ping: %w", err)
		}

		userStore = redisstore.NewUserStore(client)
		tweetStore = redisstore.NewTweetStore(client)
		followStore = redisstore.NewFollowStore(client)
		timelineStore = redisstore.NewTimelineStore(client)

	default:
		userStore = memory.NewUserStore()
		tweetStore = memory.NewTweetStore()
		followStore = memory.NewFollowStore()
		timelineStore = memory.NewTimelineStore()
	}

	fmt.Printf("storage: %s\n", cfg.Storage)

	// Service layer.
	userService := user.NewService(userStore)
	tweetService := tweet.NewService(tweetStore)
	followService := follow.NewService(followStore)
	timelineService := timeline.NewService(timelineStore, tweetService, followService)

	// Router.
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/users", handlers.CreateUser(userService))
	mux.HandleFunc("GET /v1/tweets/{id}", handlers.TweetByID(tweetService))

	authenticated := http.NewServeMux()
	authenticated.HandleFunc("POST /v1/tweets", handlers.CreateTweet(timelineService))
	authenticated.HandleFunc("POST /v1/users/{id}/follow", handlers.FollowUser(followService))
	authenticated.HandleFunc("DELETE /v1/users/{id}/follow", handlers.UnfollowUser(followService))
	authenticated.HandleFunc("GET /v1/timeline", handlers.Timeline(timelineService))

	mux.Handle("/", web.UserIDMiddleware(authenticated))

	router := web.PanicRecovery(mux)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine so we can listen for shutdown signals.
	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("twix API listening on :%s\n", cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal or server error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-quit:
		log.Printf("received signal %v, shutting down gracefully", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
