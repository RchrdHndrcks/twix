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
	"github.com/RchrdHndrcks/twix/internal/user"
)

type mockUserService struct {
	createFn func(ctx context.Context, username string) (user.User, error)
	userFn   func(ctx context.Context, id string) (user.User, error)
}

func (m *mockUserService) Create(ctx context.Context, username string) (user.User, error) {
	return m.createFn(ctx, username)
}

func (m *mockUserService) User(ctx context.Context, id string) (user.User, error) {
	return m.userFn(ctx, id)
}

func TestCreateUser(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("201 created", func(t *testing.T) {
		svc := &mockUserService{
			createFn: func(_ context.Context, username string) (user.User, error) {
				return user.User{ID: "u1", Username: username, CreatedAt: now}, nil
			},
		}

		body, _ := json.Marshal(map[string]string{"username": "alice"})
		req := httptest.NewRequest("POST", "/v1/users", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handlers.CreateUser(svc)(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		var got map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["id"] != "u1" {
			t.Errorf("expected ID u1, got %v", got["id"])
		}
		if got["username"] != "alice" {
			t.Errorf("expected Username alice, got %v", got["username"])
		}
	})

	t.Run("400 invalid body", func(t *testing.T) {
		svc := &mockUserService{}

		req := httptest.NewRequest("POST", "/v1/users", bytes.NewReader([]byte("not json")))
		rec := httptest.NewRecorder()

		handlers.CreateUser(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != "invalid request body" {
			t.Errorf("expected 'invalid request body', got %q", got["error"])
		}
	})

	t.Run("400 empty username", func(t *testing.T) {
		svc := &mockUserService{
			createFn: func(_ context.Context, _ string) (user.User, error) {
				return user.User{}, user.ErrUsernameEmpty
			},
		}

		body, _ := json.Marshal(map[string]string{"username": ""})
		req := httptest.NewRequest("POST", "/v1/users", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handlers.CreateUser(svc)(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != user.ErrUsernameEmpty.Error() {
			t.Errorf("expected %q, got %q", user.ErrUsernameEmpty.Error(), got["error"])
		}
	})

	t.Run("500 unexpected error", func(t *testing.T) {
		svc := &mockUserService{
			createFn: func(_ context.Context, _ string) (user.User, error) {
				return user.User{}, errors.New("db down")
			},
		}

		body, _ := json.Marshal(map[string]string{"username": "alice"})
		req := httptest.NewRequest("POST", "/v1/users", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handlers.CreateUser(svc)(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		var got map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&got)

		if got["error"] != "internal server error" {
			t.Errorf("expected 'internal server error', got %q", got["error"])
		}
	})
}
