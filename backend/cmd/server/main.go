package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Thruqe/quire/internal/auth"
	"github.com/Thruqe/quire/internal/db"
	"github.com/Thruqe/quire/internal/handler"
	quiremw "github.com/Thruqe/quire/internal/middleware"
)

func main() {
	ctx := context.Background()

	if err := db.Connect(ctx); err != nil {
		log.Fatalf("❌ DB connection failed: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "DB unavailable")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Quire API is running")
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"message": "pong"}`)
		})

		// Auth routes (public)
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/login", handler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(quiremw.RequireAuth)
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				claims := r.Context().Value(quiremw.UserKey).(*auth.Claims)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"user_id": "%s", "email": "%s"}`, claims.UserID, claims.Email)
			})
			// Workspaces
			r.Post("/workspaces", handler.CreateWorkspace)
			r.Get("/workspaces", handler.ListWorkspaces)
			// Pages
			r.Route("/workspaces/{workspaceID}/pages", func(r chi.Router) {
				r.Post("/", handler.CreatePage)
				r.Get("/", handler.ListPages)
			})
			r.Route("/pages/{pageID}", func(r chi.Router) {
				r.Get("/", handler.GetPage)
				r.Patch("/", handler.UpdatePage)
				r.Delete("/", handler.DeletePage)
			})
			r.Get("/pages/{pageID}/content", handler.GetPageContent)
			r.Post("/pages/{pageID}/content", handler.SavePageContent)
		})
	})

	log.Printf("🌿 Quire API running on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
