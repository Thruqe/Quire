package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"strings"

	"github.com/Thruqe/quire/internal/auth"
	"github.com/Thruqe/quire/internal/db"
	quiremw "github.com/Thruqe/quire/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type createPageRequest struct {
	Title    string  `json:"title"`
	ParentID *string `json:"parent_id"`
	Icon     *string `json:"icon"`
}

type pageResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ParentID    *string   `json:"parent_id"`
	Title       string    `json:"title"`
	Icon        *string   `json:"icon"`
	Cover       *string   `json:"cover"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreatePage(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(quiremw.UserKey).(*auth.Claims)
	workspaceID := chi.URLParam(r, "workspaceID")

	var req createPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		req.Title = "Untitled"
	}

	var page pageResponse
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO pages (workspace_id, parent_id, title, icon, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, workspace_id, parent_id, title, icon, cover, created_by, created_at, updated_at`,
		workspaceID, req.ParentID, req.Title, req.Icon, claims.UserID,
	).Scan(
		&page.ID, &page.WorkspaceID, &page.ParentID,
		&page.Title, &page.Icon, &page.Cover,
		&page.CreatedBy, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create page")
		return
	}

	writeJSON(w, http.StatusCreated, page)
}

func ListPages(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	// Returns flat list; client builds the tree
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, workspace_id, parent_id, title, icon, cover, created_by, created_at, updated_at
		 FROM pages
		 WHERE workspace_id = $1
		 ORDER BY created_at ASC`,
		workspaceID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch pages")
		return
	}
	defer rows.Close()

	pages := []pageResponse{}
	for rows.Next() {
		var p pageResponse
		if err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.ParentID,
			&p.Title, &p.Icon, &p.Cover,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			continue
		}
		pages = append(pages, p)
	}

	writeJSON(w, http.StatusOK, pages)
}

func GetPage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var page pageResponse
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, workspace_id, parent_id, title, icon, cover, created_by, created_at, updated_at
		 FROM pages WHERE id = $1`,
		pageID,
	).Scan(
		&page.ID, &page.WorkspaceID, &page.ParentID,
		&page.Title, &page.Icon, &page.Cover,
		&page.CreatedBy, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "page not found")
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func UpdatePage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var req struct {
		Title    *string `json:"title"`
		Icon     *string `json:"icon"`
		Cover    *string `json:"cover"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var page pageResponse
	err := db.Pool.QueryRow(context.Background(),
		`UPDATE pages SET
		   title     = COALESCE($1, title),
		   icon      = COALESCE($2, icon),
		   cover     = COALESCE($3, cover),
		   parent_id = COALESCE($4, parent_id),
		   updated_at = now()
		 WHERE id = $5
		 RETURNING id, workspace_id, parent_id, title, icon, cover, created_by, created_at, updated_at`,
		req.Title, req.Icon, req.Cover, req.ParentID, pageID,
	).Scan(
		&page.ID, &page.WorkspaceID, &page.ParentID,
		&page.Title, &page.Icon, &page.Cover,
		&page.CreatedBy, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update page")
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func DeletePage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	_, err := db.Pool.Exec(context.Background(),
		`DELETE FROM pages WHERE id = $1`, pageID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete page")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SavePageContent(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	_, err := db.Pool.Exec(context.Background(),
		`UPDATE pages SET properties = jsonb_set(COALESCE(properties, '{}'), '{content}', $1::jsonb), updated_at = now()
		 WHERE id = $2`,
		`"`+escapeJSON(req.Content)+`"`, pageID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save content")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func GetPageContent(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var content *string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT properties->>'content' FROM pages WHERE id = $1`,
		pageID,
	).Scan(&content)
	if err != nil {
		writeError(w, http.StatusNotFound, "page not found")
		return
	}

	if content == nil {
		writeJSON(w, http.StatusOK, map[string]string{"content": ""})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"content": *content})
}
