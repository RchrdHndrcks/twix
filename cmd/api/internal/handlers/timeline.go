package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RchrdHndrcks/twix/internal/platform/web"
)

const maxLimit = 100

// Timeline handles GET /v1/timeline.
func Timeline(svc TimelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := web.UserIDFromContext(r.Context())

		var cursor time.Time
		if c := r.URL.Query().Get("cursor"); c != "" {
			var err error
			cursor, err = time.Parse(time.RFC3339Nano, c)
			if err != nil {
				web.RespondError(w, http.StatusBadRequest, errors.New("invalid cursor format, use RFC3339"))
				return
			}
		}

		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			var err error
			limit, err = strconv.Atoi(l)
			if err != nil || limit <= 0 {
				web.RespondError(w, http.StatusBadRequest, errors.New("limit must be a positive integer"))
				return
			}
			if limit > maxLimit {
				limit = maxLimit
			}
		}

		tweets, err := svc.Timeline(r.Context(), userID, cursor, limit)
		if err != nil {
			web.RespondError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		var nextCursor string
		if len(tweets) > 0 {
			nextCursor = tweets[len(tweets)-1].CreatedAt.Format(time.RFC3339Nano)
		}

		web.RespondJSON(w, http.StatusOK, map[string]any{
			"tweets":      tweets,
			"next_cursor": nextCursor,
		})
	}
}
