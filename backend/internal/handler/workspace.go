package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Thruqe/quire/internal/auth"
	"github.com/Thruqe/quire/internal/db"
	quiremw "github.com/Thruqe/quire/internal/middleware"
)

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

type workspaceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(quiremw.UserKey).(*auth.Claims)

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var ws workspaceResponse
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO workspaces (name, owner_id)
		 VALUES ($1, $2)
		 RETURNING id, name, owner_id, created_at`,
		req.Name, claims.UserID,
	).Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}

	writeJSON(w, http.StatusCreated, ws)
}

func ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(quiremw.UserKey).(*auth.Claims)

	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, name, owner_id, created_at
		 FROM workspaces
		 WHERE owner_id = $1
		 ORDER BY created_at ASC`,
		claims.UserID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch workspaces")
		return
	}
	defer rows.Close()

	workspaces := []workspaceResponse{}
	for rows.Next() {
		var ws workspaceResponse
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.CreatedAt); err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}

	writeJSON(w, http.StatusOK, workspaces)
}
