package utils

import (
	"encoding/json"
	"log"
	"mockgitea/internal/config"
	"net/http"
	"strconv"
	"strings"
)

func EnsureGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "method not allowed"})
		return false
	}
	return true
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[mock-gitea] write json failed: %v", err)
	}
}

func ParsePositiveInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func Paginate[T any](items []T, page, limit int) []T {
	if limit <= 0 {
		limit = config.DefaultLimit
	}
	if page <= 0 {
		page = config.DefaultPage
	}

	start := (page - 1) * limit
	if start >= len(items) {
		return []T{}
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	return append([]T(nil), items[start:end]...)
}
