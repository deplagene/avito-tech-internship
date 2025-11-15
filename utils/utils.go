package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func ParseJson(r *http.Request, v any) error {
	const op = "utils.ParseJson"

	if r.Body == nil {
		return fmt.Errorf("%s: body is nil", op)
	}

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	const op = "utils.WriteJson"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func WriteError(w http.ResponseWriter, logger *slog.Logger, status int, err error) {
	const op = "utils.WriteError"

	logger.Error("request failed", "status", status, Err(err))

	if status >= 500 {
		msg := map[string]string{"error": "internal server error"}

		if err := WriteJson(w, status, msg); err != nil {
			logger.Error("failed to write 5xx response", Err(err))
		}

		return
	}

	msg := map[string]string{"error": err.Error()}
	if err := WriteJson(w, status, msg); err != nil {
		logger.Error("failed to write 4xx response", Err(err))

	}
}
