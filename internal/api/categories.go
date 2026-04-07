package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type categoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Emoji       string `json:"emoji"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	IsSystem    bool   `json:"is_system"`
	IsProtected bool   `json:"is_protected"`
}

type categoriesListResponse struct {
	Categories []categoryResponse `json:"categories"`
}

type createCategoryRequest struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

type updateCategoryRequest struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

// categoriesHandler handles CRUD for /api/v1/categories[/{id}].
func categoriesHandler(catSvc *service.CategoryService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/categories")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			id, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid category id"})
				return
			}

			switch r.Method {
			case http.MethodPut:
				var req updateCategoryRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
					return
				}
				cat, err := catSvc.Update(ctx, userID, id, req.Name, req.Emoji, req.Type, req.Color)
				if err != nil {
					writeError(w, log, err)
					return
				}
				writeJSON(w, http.StatusOK, toCategoryResponse(cat))
			case http.MethodDelete:
				if err := catSvc.Delete(ctx, userID, id); err != nil {
					writeError(w, log, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			catType := r.URL.Query().Get("type")  // "expense" | "income" | "both" | ""
			order := r.URL.Query().Get("order")    // "asc" | "desc" | ""

			list, err := catSvc.ListSorted(ctx, userID, catType, order)
			if err != nil {
				if errors.Is(err, domain.ErrInvalidSortParam) {
					writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
					return
				}
				writeError(w, log, err)
				return
			}
			items := make([]categoryResponse, 0, len(list))
			for _, c := range list {
				items = append(items, toCategoryResponse(c))
			}
			writeJSON(w, http.StatusOK, categoriesListResponse{Categories: items})

		case http.MethodPost:
			var req createCategoryRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
				return
			}
			cat, err := catSvc.Create(ctx, userID, req.Name, req.Emoji, req.Type, req.Color)
			if err != nil {
				writeError(w, log, err)
				return
			}
			writeJSON(w, http.StatusCreated, toCategoryResponse(cat))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func toCategoryResponse(c *domain.Category) categoryResponse {
	return categoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Emoji:       c.Emoji,
		Type:        string(c.Type),
		Color:       c.Color,
		IsSystem:    c.IsSystem(),
		IsProtected: c.IsProtected,
	}
}
