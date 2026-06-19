package server

import (
	"encoding/json"
	"net/http"
)

// writeError writes a JSON {"error": message} response with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// decodeJSON decodes a JSON request body into v. It reports a 400 to the client
// automatically on failure and returns false so callers can early-return.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "request body is required")
		return false
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
