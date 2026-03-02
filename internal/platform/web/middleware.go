package web

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
)

type contextKey string

const userIDKey contextKey = "userID"

// ContextWithUserID returns a new context with the given user ID.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

// UserIDMiddleware extracts the X-User-ID header and injects it into the context.
// Returns 401 if the header is missing.
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			RespondError(w, http.StatusUnauthorized, errors.New("X-User-ID header is required"))
			return
		}

		ctx := ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PanicRecovery catches panics from downstream handlers and returns a 500
// response instead of crashing the server.
func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
