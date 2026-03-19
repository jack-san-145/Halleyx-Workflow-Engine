package api

import (
	"context"
	"encoding/json"
	"halleyx-workflow-docker/internal/engine"
	"halleyx-workflow-docker/internal/store"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExecuteRequest struct {
	TriggeredBy string                 `json:"triggered_by"`
	Data        map[string]interface{} `json:"data"`
}

type ApprovalRequest struct {
	ApproverID string                 `json:"approver_id"`
	Data       map[string]interface{} `json:"data"`
}

func ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowID, _ := uuid.Parse(chi.URLParam(r, "workflowID"))
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	workflow, err := store.GetWorkflow(context.Background(), workflowID)
	if err != nil {
		http.Error(w, "workflow not found", http.StatusNotFound)
		return
	}

	exec := &store.Execution{
		ID:              uuid.New(),
		WorkflowID:      workflowID,
		WorkflowVersion: workflow.Version,
		Status:          "pending",
		Data:            req.Data,
		CurrentStepID:   workflow.StartStepID,
	}
	if req.TriggeredBy != "" {
		if tid, err := uuid.Parse(req.TriggeredBy); err == nil {
			exec.TriggeredBy = &tid
		}
	}

	if err := store.InsertExecution(context.Background(), exec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go engine.RunExecution(exec.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func GetExecution(w http.ResponseWriter, r *http.Request) {
	executionID, _ := uuid.Parse(chi.URLParam(r, "executionID"))
	exec, err := store.GetExecution(context.Background(), executionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	logs, _ := store.ListExecutionLogs(context.Background(), executionID)

	resp := map[string]interface{}{
		"id":              exec.ID,
		"workflow_id":     exec.WorkflowID,
		"workflow_name":   store.GetWorkflowName(exec.WorkflowID),
		"status":          exec.Status,
		"current_step_id": exec.CurrentStepID,
		"data":            exec.Data,
		"started_at":      exec.StartedAt,
		"ended_at":        exec.EndedAt,
		"retries":         exec.Retries,
		"logs":            logs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func RetryExecution(w http.ResponseWriter, r *http.Request) {
	executionID, _ := uuid.Parse(chi.URLParam(r, "executionID"))
	exec, err := store.GetExecution(context.Background(), executionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	exec.Retries++
	exec.Status = "pending"
	exec.EndedAt = nil
	store.UpdateExecution(context.Background(), exec)

	go engine.RunExecution(exec.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func ApproveExecution(w http.ResponseWriter, r *http.Request) {
	executionID, _ := uuid.Parse(chi.URLParam(r, "executionID"))
	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exec, err := store.GetExecution(context.Background(), executionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if exec.Status != "pending" {
		http.Error(w, "execution is not pending approval", http.StatusBadRequest)
		return
	}

	// Merge data (approval context)
	if exec.Data == nil {
		exec.Data = map[string]interface{}{}
	}
	for k, v := range req.Data {
		exec.Data[k] = v
	}

	exec.Status = "in_progress"
	store.UpdateExecution(context.Background(), exec)
	go engine.RunExecution(exec.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func CancelExecution(w http.ResponseWriter, r *http.Request) {
	executionID, _ := uuid.Parse(chi.URLParam(r, "executionID"))
	store.MarkExecutionCanceled(context.Background(), executionID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "canceled"})
}
