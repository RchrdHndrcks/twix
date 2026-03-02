package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RchrdHndrcks/twix/cmd/api/internal/handlers"
	"github.com/RchrdHndrcks/twix/internal/follow"
)

type mockFollowService struct {
	followFn   func(ctx context.Context, followerID, followeeID string) error
	unfollowFn func(ctx context.Context, followerID, followeeID string) error
}

func (m *mockFollowService) Follow(ctx context.Context, followerID, followeeID string) error {
	return m.followFn(ctx, followerID, followeeID)
}

func (m *mockFollowService) Unfollow(ctx context.Context, followerID, followeeID string) error {
	return m.unfollowFn(ctx, followerID, followeeID)
}

func TestFollowUser(t *testing.T) {
	t.Run("204 success", func(t *testing.T) {
		svc := &mockFollowService{
			followFn: func(_ context.Context, _, _ string) error {
				return nil
			},
		}

		req := newAuthenticatedRequest("POST", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.FollowUser(svc)(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("400 self follow", func(t *testing.T) {
		svc := &mockFollowService{
			followFn: func(_ context.Context, _, _ string) error {
				return follow.ErrSelfFollow
			},
		}

		req := newAuthenticatedRequest("POST", "/v1/users/u1/follow", nil, "u1")
		req.SetPathValue("id", "u1")
		rec := httptest.NewRecorder()

		handlers.FollowUser(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != follow.ErrSelfFollow.Error() {
			t.Errorf("expected %q, got %q", follow.ErrSelfFollow.Error(), got["error"])
		}
	})

	t.Run("409 already following", func(t *testing.T) {
		svc := &mockFollowService{
			followFn: func(_ context.Context, _, _ string) error {
				return follow.ErrAlreadyFollowing
			},
		}

		req := newAuthenticatedRequest("POST", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.FollowUser(svc)(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != follow.ErrAlreadyFollowing.Error() {
			t.Errorf("expected %q, got %q", follow.ErrAlreadyFollowing.Error(), got["error"])
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockFollowService{
			followFn: func(_ context.Context, _, _ string) error {
				return errors.New("db error")
			},
		}

		req := newAuthenticatedRequest("POST", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.FollowUser(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestUnfollowUser(t *testing.T) {
	t.Run("204 success", func(t *testing.T) {
		svc := &mockFollowService{
			unfollowFn: func(_ context.Context, _, _ string) error {
				return nil
			},
		}

		req := newAuthenticatedRequest("DELETE", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.UnfollowUser(svc)(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("404 not following", func(t *testing.T) {
		svc := &mockFollowService{
			unfollowFn: func(_ context.Context, _, _ string) error {
				return follow.ErrNotFollowing
			},
		}

		req := newAuthenticatedRequest("DELETE", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.UnfollowUser(svc)(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != follow.ErrNotFollowing.Error() {
			t.Errorf("expected %q, got %q", follow.ErrNotFollowing.Error(), got["error"])
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockFollowService{
			unfollowFn: func(_ context.Context, _, _ string) error {
				return errors.New("db error")
			},
		}

		req := newAuthenticatedRequest("DELETE", "/v1/users/u2/follow", nil, "u1")
		req.SetPathValue("id", "u2")
		rec := httptest.NewRecorder()

		handlers.UnfollowUser(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}
