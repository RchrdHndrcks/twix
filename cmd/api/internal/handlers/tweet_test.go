package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/cmd/api/internal/handlers"
	"github.com/RchrdHndrcks/twix/internal/platform/web"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

type mockTimelineService struct {
	publishTweetFn func(ctx context.Context, authorID, content string) (tweet.Tweet, error)
	timelineFn     func(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error)
}

func (m *mockTimelineService) PublishTweet(ctx context.Context, authorID, content string) (tweet.Tweet, error) {
	return m.publishTweetFn(ctx, authorID, content)
}

func (m *mockTimelineService) Timeline(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error) {
	return m.timelineFn(ctx, userID, cursor, limit)
}

type mockTweetReader struct {
	tweetFn func(ctx context.Context, id string) (tweet.Tweet, error)
}

func (m *mockTweetReader) Tweet(ctx context.Context, id string) (tweet.Tweet, error) {
	return m.tweetFn(ctx, id)
}

func newAuthenticatedRequest(method, target string, body []byte, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	ctx := web.ContextWithUserID(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestCreateTweet(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("201 created", func(t *testing.T) {
		svc := &mockTimelineService{
			publishTweetFn: func(_ context.Context, authorID, content string) (tweet.Tweet, error) {
				return tweet.Tweet{ID: "t1", AuthorID: authorID, Content: content, CreatedAt: now}, nil
			},
		}

		body, _ := json.Marshal(map[string]string{"content": "hello world"})
		req := newAuthenticatedRequest("POST", "/v1/tweets", body, "u1")
		rec := httptest.NewRecorder()

		handlers.CreateTweet(svc)(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		var got map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["id"] != "t1" {
			t.Errorf("expected ID t1, got %v", got["id"])
		}
		if got["content"] != "hello world" {
			t.Errorf("expected Content 'hello world', got %v", got["content"])
		}
	})

	t.Run("400 invalid body", func(t *testing.T) {
		svc := &mockTimelineService{}

		req := newAuthenticatedRequest("POST", "/v1/tweets", []byte("bad"), "u1")
		rec := httptest.NewRecorder()

		handlers.CreateTweet(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != "invalid request body" {
			t.Errorf("expected 'invalid request body', got %q", got["error"])
		}
	})

	t.Run("400 empty content", func(t *testing.T) {
		svc := &mockTimelineService{
			publishTweetFn: func(_ context.Context, _, _ string) (tweet.Tweet, error) {
				return tweet.Tweet{}, tweet.ErrContentEmpty
			},
		}

		body, _ := json.Marshal(map[string]string{"content": ""})
		req := newAuthenticatedRequest("POST", "/v1/tweets", body, "u1")
		rec := httptest.NewRecorder()

		handlers.CreateTweet(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != tweet.ErrContentEmpty.Error() {
			t.Errorf("expected %q, got %q", tweet.ErrContentEmpty.Error(), got["error"])
		}
	})

	t.Run("400 content too long", func(t *testing.T) {
		svc := &mockTimelineService{
			publishTweetFn: func(_ context.Context, _, _ string) (tweet.Tweet, error) {
				return tweet.Tweet{}, tweet.ErrContentTooLong
			},
		}

		body, _ := json.Marshal(map[string]string{"content": "x"})
		req := newAuthenticatedRequest("POST", "/v1/tweets", body, "u1")
		rec := httptest.NewRecorder()

		handlers.CreateTweet(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != tweet.ErrContentTooLong.Error() {
			t.Errorf("expected %q, got %q", tweet.ErrContentTooLong.Error(), got["error"])
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockTimelineService{
			publishTweetFn: func(_ context.Context, _, _ string) (tweet.Tweet, error) {
				return tweet.Tweet{}, errors.New("storage failure")
			},
		}

		body, _ := json.Marshal(map[string]string{"content": "hello"})
		req := newAuthenticatedRequest("POST", "/v1/tweets", body, "u1")
		rec := httptest.NewRecorder()

		handlers.CreateTweet(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestTweetByID(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("200 found", func(t *testing.T) {
		svc := &mockTweetReader{
			tweetFn: func(_ context.Context, id string) (tweet.Tweet, error) {
				return tweet.Tweet{ID: id, AuthorID: "u1", Content: "hello", CreatedAt: now}, nil
			},
		}

		req := httptest.NewRequest("GET", "/v1/tweets/t1", nil)
		req.SetPathValue("id", "t1")
		rec := httptest.NewRecorder()

		handlers.TweetByID(svc)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var got map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["id"] != "t1" {
			t.Errorf("expected ID t1, got %v", got["id"])
		}
		if got["content"] != "hello" {
			t.Errorf("expected Content 'hello', got %v", got["content"])
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		svc := &mockTweetReader{
			tweetFn: func(_ context.Context, _ string) (tweet.Tweet, error) {
				return tweet.Tweet{}, tweet.ErrNotFound
			},
		}

		req := httptest.NewRequest("GET", "/v1/tweets/missing", nil)
		req.SetPathValue("id", "missing")
		rec := httptest.NewRecorder()

		handlers.TweetByID(svc)(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != tweet.ErrNotFound.Error() {
			t.Errorf("expected %q, got %q", tweet.ErrNotFound.Error(), got["error"])
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockTweetReader{
			tweetFn: func(_ context.Context, _ string) (tweet.Tweet, error) {
				return tweet.Tweet{}, errors.New("boom")
			},
		}

		req := httptest.NewRequest("GET", "/v1/tweets/t1", nil)
		req.SetPathValue("id", "t1")
		rec := httptest.NewRecorder()

		handlers.TweetByID(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}
