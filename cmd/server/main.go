package main

import (
	"fmt"
	"halleyx-workflow-docker/internal/api"
	"halleyx-workflow-docker/internal/store"
	"halleyx-workflow-docker/internal/ws"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Database configuration
	dsn := os.Getenv("DATABASE_URL")

	if err := store.ConnectPostgres(dsn); err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}

	// Start WebSocket hub
	ws.Start()

	r := chi.NewRouter()

	// Setup router

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	// UI pages
	r.Get("/", api.ServeIndex)
	r.Get("/index_page", api.ServeIndex)
	r.Get("/execute_page", api.ServeExecute)
	r.Get("/execution_page", api.ServeExecution)
	r.Get("/workflow_detail_page", api.ServeWorkflowDetail)

	// API
	// Workflows
	r.Post("/workflows", api.CreateWorkflow)
	r.Get("/workflows", api.ListWorkflows)
	r.Get("/workflows/{id}", api.GetWorkflow)

	// Steps
	r.Post("/workflows/{workflowID}/steps", api.CreateStep)
	r.Get("/workflows/{workflowID}/steps", api.ListSteps)

	// Rules
	r.Post("/steps/{stepID}/rules", api.CreateRule)
	r.Get("/steps/{stepID}/rules", api.ListRules)

	// Executions
	r.Post("/workflows/{workflowID}/execute", api.ExecuteWorkflow)
	r.Get("/executions/{executionID}", api.GetExecution)
	r.Post("/executions/{executionID}/retry", api.RetryExecution)
	r.Post("/executions/{executionID}/cancel", api.CancelExecution)
	r.Post("/executions/{executionID}/approve", api.ApproveExecution)

	// WebSockets
	r.Get("/ws", api.WebSocketHandler)

	addr := ":8080"
	fmt.Println("Server running at", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
