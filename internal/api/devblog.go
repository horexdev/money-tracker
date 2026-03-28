package api

import (
	"log/slog"
	"net/http"
	"strings"

	devblog "github.com/horexdev/money-tracker/devblogs"
)

type devblogEntryResponse struct {
	Filename string `json:"filename"`
	Version  string `json:"version"`
	Date     string `json:"date"`
}

type devblogListResponse struct {
	Entries []devblogEntryResponse `json:"entries"`
}

type devblogContentResponse struct {
	Filename string `json:"filename"`
	Version  string `json:"version"`
	Content  string `json:"content"`
}

// devblogListHandler handles GET /api/v1/devblog
func devblogListHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		entries, err := devblog.List()
		if err != nil {
			log.ErrorContext(r.Context(), "devblog: list failed", slog.String("error", err.Error()))
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load devblog"})
			return
		}

		items := make([]devblogEntryResponse, 0, len(entries))
		for _, e := range entries {
			items = append(items, devblogEntryResponse{
				Filename: e.Filename,
				Version:  e.Version,
				Date:     e.Date.Format("2006-01-02"),
			})
		}
		writeJSON(w, http.StatusOK, devblogListResponse{Entries: items})
	}
}

// devblogEntryHandler handles GET /api/v1/devblog/{filename}
func devblogEntryHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract filename from path: /api/v1/devblog/{filename}
		filename := strings.TrimPrefix(r.URL.Path, "/api/v1/devblog/")
		if filename == "" || strings.Contains(filename, "/") {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid filename"})
			return
		}

		raw, err := devblog.Content(filename)
		if err != nil {
			log.ErrorContext(r.Context(), "devblog: read failed", slog.String("filename", filename), slog.String("error", err.Error()))
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "devblog entry not found"})
			return
		}

		// Parse version from filename for the response.
		version := ""
		if idx := strings.Index(filename, "_"); idx > 0 {
			version = filename[:idx]
		}

		writeJSON(w, http.StatusOK, devblogContentResponse{
			Filename: filename,
			Version:  version,
			Content:  string(raw),
		})
	}
}
