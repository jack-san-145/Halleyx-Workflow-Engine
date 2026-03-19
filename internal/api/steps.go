package api

import (
	"context"
	"encoding/json"
	"halleyx-workflow-docker/internal/store"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CreateStep(w http.ResponseWriter, r *http.Request) {
	workflowID, _ := uuid.Parse(chi.URLParam(r, "workflowID"))
	var step store.Step
	if err := json.NewDecoder(r.Body).Decode(&step); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	step.ID = uuid.New()
	step.WorkflowID = workflowID
	if err := store.InsertStep(context.Background(), &step); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If workflow has no start step yet, set this as the start
	wf, err := store.GetWorkflow(context.Background(), workflowID)
	if err == nil && wf != nil && wf.StartStepID == nil {
		_ = store.UpdateWorkflowStartStep(context.Background(), workflowID, step.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(step)
}

func ListSteps(w http.ResponseWriter, r *http.Request) {
	workflowID, _ := uuid.Parse(chi.URLParam(r, "workflowID"))
	steps, err := store.ListSteps(context.Background(), workflowID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(steps)
}

// Rules
func CreateRule(w http.ResponseWriter, r *http.Request) {
	stepID, _ := uuid.Parse(chi.URLParam(r, "stepID"))
	var rule store.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rule.ID = uuid.New()
	rule.StepID = stepID
	if err := store.InsertRule(context.Background(), &rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rule)
}

func ListRules(w http.ResponseWriter, r *http.Request) {
	stepID, _ := uuid.Parse(chi.URLParam(r, "stepID"))
	rules, err := store.ListRules(context.Background(), stepID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}
