package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/RchrdHndrcks/twix/internal/follow"
	"github.com/RchrdHndrcks/twix/internal/platform/web"
)

// FollowService defines the contract the follow handlers depend on.
type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
}

// FollowUser handles POST /v1/users/{id}/follow.
func FollowUser(svc FollowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := web.UserIDFromContext(r.Context())
		followeeID := r.PathValue("id")

		if err := svc.Follow(r.Context(), userID, followeeID); err != nil {
			if errors.Is(err, follow.ErrSelfFollow) {
				web.RespondError(w, http.StatusBadRequest, follow.ErrSelfFollow)
				return
			}
			if errors.Is(err, follow.ErrAlreadyFollowing) {
				web.RespondError(w, http.StatusConflict, follow.ErrAlreadyFollowing)
				return
			}
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UnfollowUser handles DELETE /v1/users/{id}/follow.
func UnfollowUser(svc FollowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := web.UserIDFromContext(r.Context())
		followeeID := r.PathValue("id")

		if err := svc.Unfollow(r.Context(), userID, followeeID); err != nil {
			if errors.Is(err, follow.ErrNotFollowing) {
				web.RespondError(w, http.StatusNotFound, follow.ErrNotFollowing)
				return
			}
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
