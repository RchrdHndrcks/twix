package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RchrdHndrcks/twix/cmd/api/internal/handlers"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

func TestTimeline(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("200 with tweets", func(t *testing.T) {
		svc := &mockTimelineService{
			timelineFn: func(_ context.Context, _ string, _ time.Time, _ int) ([]tweet.Tweet, error) {
				return []tweet.Tweet{
					{ID: "t1", AuthorID: "u2", Content: "hello", CreatedAt: now},
					{ID: "t2", AuthorID: "u2", Content: "world", CreatedAt: now.Add(-time.Second)},
				}, nil
			},
		}

		req := newAuthenticatedRequest("GET", "/v1/timeline", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var got struct {
			Tweets     []map[string]any `json:"tweets"`
			NextCursor string           `json:"next_cursor"`
		}
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if len(got.Tweets) != 2 {
			t.Fatalf("expected 2 tweets, got %d", len(got.Tweets))
		}
		if got.NextCursor == "" {
			t.Error("expected non-empty next_cursor")
		}
	})

	t.Run("200 empty timeline", func(t *testing.T) {
		svc := &mockTimelineService{
			timelineFn: func(_ context.Context, _ string, _ time.Time, _ int) ([]tweet.Tweet, error) {
				return nil, nil
			},
		}

		req := newAuthenticatedRequest("GET", "/v1/timeline", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var got struct {
			Tweets     []map[string]any `json:"tweets"`
			NextCursor string           `json:"next_cursor"`
		}
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got.NextCursor != "" {
			t.Errorf("expected empty next_cursor, got %q", got.NextCursor)
		}
	})

	t.Run("400 invalid cursor", func(t *testing.T) {
		svc := &mockTimelineService{}

		req := newAuthenticatedRequest("GET", "/v1/timeline?cursor=not-a-date", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != "invalid cursor format, use RFC3339" {
			t.Errorf("expected cursor error, got %q", got["error"])
		}
	})

	t.Run("400 invalid limit", func(t *testing.T) {
		svc := &mockTimelineService{}

		req := newAuthenticatedRequest("GET", "/v1/timeline?limit=abc", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != "limit must be a positive integer" {
			t.Errorf("expected limit error, got %q", got["error"])
		}
	})

	t.Run("400 negative limit", func(t *testing.T) {
		svc := &mockTimelineService{}

		req := newAuthenticatedRequest("GET", "/v1/timeline?limit=-1", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockTimelineService{
			timelineFn: func(_ context.Context, _ string, _ time.Time, _ int) ([]tweet.Tweet, error) {
				return nil, errors.New("storage error")
			},
		}

		req := newAuthenticatedRequest("GET", "/v1/timeline", nil, "u1")
		rec := httptest.NewRecorder()

		handlers.Timeline(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}
