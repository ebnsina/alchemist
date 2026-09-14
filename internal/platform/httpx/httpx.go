// Package httpx holds the response conventions every module shares.
package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON writes a success response.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a failure as a stable machine code plus plain-language text.
//
// The code is the contract: customers branch on it and it must not change. The
// message is for a human and may be reworded or translated freely. Raw database or
// ffmpeg text never reaches either field.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}
