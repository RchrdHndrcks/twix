package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/RchrdHndrcks/twix/cmd/api/internal/handlers"
	redisstore "github.com/RchrdHndrcks/twix/internal/platform/storage/redis"
	"github.com/RchrdHndrcks/twix/internal/platform/web"
	"github.com/RchrdHndrcks/twix/internal/timeline"
	"github.com/RchrdHndrcks/twix/internal/tweet"
	"github.com/RchrdHndrcks/twix/internal/user"
	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

type testServer struct {
	baseURL string
	client  *http.Client
}

// tryStartRedis attempts to start a Redis container, recovering from panics
// caused by Docker not being available (testcontainers panics internally).
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

func setupServer(t *testing.T) *testServer {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := tryStartRedis(t, ctx)
	if err != nil {
		t.Skipf("skipping: Docker not available (%v)", err)
	}

	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}

	opts, err := redis.ParseURL(connStr)
	if err != nil {
		t.Fatalf("failed to parse redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	// Wire up the full stack with Redis stores.
	userStore := redisstore.NewUserStore(client)
	tweetStore := redisstore.NewTweetStore(client)
	followStore := redisstore.NewFollowStore(client)
	timelineStore := redisstore.NewTimelineStore(client)

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

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := &http.Server{Handler: router}
	go func() { _ = server.Serve(listener) }()

	t.Cleanup(func() {
		_ = server.Close()
	})

	return &testServer{
		baseURL: fmt.Sprintf("http://%s", listener.Addr().String()),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (ts *testServer) createUser(t *testing.T, username string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"username": username})
	resp, err := ts.client.Post(ts.baseURL+"/v1/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.ID
}

func (ts *testServer) publishTweet(t *testing.T, userID, content string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"content": content})
	req, _ := http.NewRequest("POST", ts.baseURL+"/v1/tweets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("failed to publish tweet: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.ID
}

func (ts *testServer) followUser(t *testing.T, followerID, followeeID string) int {
	t.Helper()

	req, _ := http.NewRequest("POST", ts.baseURL+"/v1/users/"+followeeID+"/follow", nil)
	req.Header.Set("X-User-ID", followerID)

	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("failed to follow user: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func (ts *testServer) timeline(t *testing.T, userID string) []map[string]any {
	t.Helper()

	req, _ := http.NewRequest("GET", ts.baseURL+"/v1/timeline", nil)
	req.Header.Set("X-User-ID", userID)

	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result struct {
		Tweets     []map[string]any `json:"tweets"`
		NextCursor string           `json:"next_cursor"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Tweets
}

func TestIntegration_FullFlow(t *testing.T) {
	ts := setupServer(t)

	alice := ts.createUser(t, "alice")
	bob := ts.createUser(t, "bob")

	status := ts.followUser(t, alice, bob)
	if status != http.StatusNoContent {
		t.Fatalf("expected 204 on follow, got %d", status)
	}

	tweetID := ts.publishTweet(t, bob, "Hello from Bob")
	if tweetID == "" {
		t.Fatal("expected non-empty tweet ID")
	}

	tweets := ts.timeline(t, alice)
	if len(tweets) != 1 {
		t.Fatalf("expected 1 tweet in Alice's timeline, got %d", len(tweets))
	}

	if tweets[0]["content"] != "Hello from Bob" {
		t.Errorf("expected 'Hello from Bob', got %s", tweets[0]["content"])
	}

	bobTweets := ts.timeline(t, bob)
	if len(bobTweets) != 0 {
		t.Errorf("expected 0 tweets in Bob's timeline, got %d", len(bobTweets))
	}
}

func TestIntegration_SelfFollow(t *testing.T) {
	ts := setupServer(t)

	alice := ts.createUser(t, "alice")

	req, _ := http.NewRequest("POST", ts.baseURL+"/v1/users/"+alice+"/follow", nil)
	req.Header.Set("X-User-ID", alice)

	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 on self-follow, got %d", resp.StatusCode)
	}
}

func TestIntegration_TweetValidation(t *testing.T) {
	ts := setupServer(t)

	alice := ts.createUser(t, "alice")

	t.Run("empty content", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"content": ""})
		req, _ := http.NewRequest("POST", ts.baseURL+"/v1/tweets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", alice)

		resp, err := ts.client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("missing X-User-ID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"content": "test"})
		resp, err := ts.client.Post(ts.baseURL+"/v1/tweets", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

func TestIntegration_GetTweetByID(t *testing.T) {
	ts := setupServer(t)

	bob := ts.createUser(t, "bob")
	tweetID := ts.publishTweet(t, bob, "A specific tweet")

	resp, err := ts.client.Get(ts.baseURL + "/v1/tweets/" + tweetID)
	if err != nil {
		t.Fatalf("failed to get tweet: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var tw map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&tw)

	if tw["content"] != "A specific tweet" {
		t.Errorf("expected 'A specific tweet', got %s", tw["content"])
	}
}

func TestIntegration_TimelinePagination(t *testing.T) {
	ts := setupServer(t)

	alice := ts.createUser(t, "alice")
	bob := ts.createUser(t, "bob")

	ts.followUser(t, alice, bob)

	for i := 0; i < 5; i++ {
		ts.publishTweet(t, bob, fmt.Sprintf("Tweet %d", i+1))
		time.Sleep(10 * time.Millisecond)
	}

	req, _ := http.NewRequest("GET", ts.baseURL+"/v1/timeline?limit=3", nil)
	req.Header.Set("X-User-ID", alice)

	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var page1 struct {
		Tweets     []map[string]any `json:"tweets"`
		NextCursor string           `json:"next_cursor"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&page1)

	if len(page1.Tweets) != 3 {
		t.Fatalf("expected 3 tweets on page 1, got %d", len(page1.Tweets))
	}

	if page1.NextCursor == "" {
		t.Fatal("expected non-empty next_cursor")
	}

	req2, _ := http.NewRequest("GET", ts.baseURL+"/v1/timeline?limit=3&cursor="+page1.NextCursor, nil)
	req2.Header.Set("X-User-ID", alice)

	resp2, err := ts.client.Do(req2)
	if err != nil {
		t.Fatalf("failed to get timeline page 2: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	var page2 struct {
		Tweets []map[string]any `json:"tweets"`
	}
	_ = json.NewDecoder(resp2.Body).Decode(&page2)

	if len(page2.Tweets) != 2 {
		t.Fatalf("expected 2 tweets on page 2, got %d", len(page2.Tweets))
	}

	seen := make(map[string]bool)
	for _, tw := range page1.Tweets {
		seen[tw["id"].(string)] = true
	}
	for _, tw := range page2.Tweets {
		if seen[tw["id"].(string)] {
			t.Errorf("duplicate tweet found across pages: %s", tw["id"])
		}
	}
}
