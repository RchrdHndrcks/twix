// Package web provides HTTP infrastructure: middleware and response helpers.
package web

import (
	"encoding/json"
	"net/http"
)

// RespondJSON writes a JSON response with the given status code and payload.
//
// Panics if encoding fails because the status code has already been written
// to the wire and there is no way to communicate the error to the client.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// At this point the status code has already been sent to the client, so we can't
			// write an error response. The best we can do is log the error and panic to stop
			// the handler.
			panic("web: failed to encode JSON response: " + err.Error())
		}
	}
}

// RespondError writes a JSON error response with the given status code and error message.
//
// It will panic if the error is nil, since that would indicate a programming error where
// the caller forgot to provide an error message. The error message is included in the
// response body as a JSON object with an "error" field.
func RespondError(w http.ResponseWriter, status int, err error) {
	if err == nil {
		// Panic if the error is nil, since that would indicate an invalid use
		// of this function where the caller forgot to provide an error message.
		panic("web: RespondError called with nil error")
	}

	RespondJSON(w, status, map[string]string{"error": err.Error()})
}
