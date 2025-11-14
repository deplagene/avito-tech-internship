package utils

import (
	"encoding/json"
	"fmt"
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

func WriteError(w http.ResponseWriter, status int, err error) {

}
