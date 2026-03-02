package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/RchrdHndrcks/twix/internal/platform/web"
	"github.com/RchrdHndrcks/twix/internal/tweet"
)

// TimelineService defines the contract the tweet and timeline handlers depend on.
type TimelineService interface {
	PublishTweet(ctx context.Context, authorID, content string) (tweet.Tweet, error)
	Timeline(ctx context.Context, userID string, cursor time.Time, limit int) ([]tweet.Tweet, error)
}

// TweetReader defines the contract for retrieving a single tweet.
type TweetReader interface {
	Tweet(ctx context.Context, id string) (tweet.Tweet, error)
}

// CreateTweet handles POST /v1/tweets.
func CreateTweet(svc TimelineService) http.HandlerFunc {
	type request struct {
		Content string `json:"content"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := web.UserIDFromContext(r.Context())

		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.RespondError(w, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		tw, err := svc.PublishTweet(r.Context(), userID, req.Content)
		if err != nil {
			if errors.Is(err, tweet.ErrContentEmpty) {
				web.RespondError(w, http.StatusBadRequest, tweet.ErrContentEmpty)
				return
			}
			if errors.Is(err, tweet.ErrContentTooLong) {
				web.RespondError(w, http.StatusBadRequest, tweet.ErrContentTooLong)
				return
			}
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		web.RespondJSON(w, http.StatusCreated, tw)
	}
}

// TweetByID handles GET /v1/tweets/{id}.
func TweetByID(svc TweetReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		tw, err := svc.Tweet(r.Context(), id)
		if err != nil {
			if errors.Is(err, tweet.ErrNotFound) {
				web.RespondError(w, http.StatusNotFound, tweet.ErrNotFound)
				return
			}
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		web.RespondJSON(w, http.StatusOK, tw)
	}
}
