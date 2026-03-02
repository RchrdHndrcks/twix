// Package handlers provides HTTP handler functions for the Twix API.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RchrdHndrcks/twix/internal/platform/web"
	"github.com/RchrdHndrcks/twix/internal/user"
)

// UserService defines the contract the user handlers depend on.
type UserService interface {
	Create(ctx context.Context, username string) (user.User, error)
	User(ctx context.Context, id string) (user.User, error)
}

// CreateUser handles POST /v1/users.
func CreateUser(svc UserService) http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.RespondError(w, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		usr, err := svc.Create(r.Context(), req.Username)
		if err != nil {
			if errors.Is(err, user.ErrUsernameEmpty) {
				web.RespondError(w, http.StatusBadRequest, user.ErrUsernameEmpty)
				return
			}
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		web.RespondJSON(w, http.StatusCreated, usr)
	}
}
