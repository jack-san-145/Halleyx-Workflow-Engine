package store

import (
	"context"

	"github.com/google/uuid"
)

type Workflow struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	IsActive    bool                   `json:"is_active"`
	InputSchema map[string]interface{} `json:"input_schema"`
	StartStepID *uuid.UUID             `json:"start_step_id"`
}

// Insert workflow
func InsertWorkflow(ctx context.Context, wf *Workflow) error {
	if wf.Version == 0 {
		wf.Version = 1
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO workflows(id,name,version,is_active,input_schema,start_step_id,created_at,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
		wf.ID, wf.Name, wf.Version, wf.IsActive, wf.InputSchema, wf.StartStepID,
	)
	return err
}

// Get workflow by ID
func GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	row := pool.QueryRow(ctx, `SELECT id,name,version,is_active,input_schema,start_step_id FROM workflows WHERE id=$1`, id)
	wf := &Workflow{}
	err := row.Scan(&wf.ID, &wf.Name, &wf.Version, &wf.IsActive, &wf.InputSchema, &wf.StartStepID)
	return wf, err
}

// List all workflows
func ListWorkflows(ctx context.Context) ([]*Workflow, error) {
	rows, err := pool.Query(ctx, `SELECT id,name,version,is_active,input_schema,start_step_id FROM workflows`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Workflow
	for rows.Next() {
		wf := &Workflow{}
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.Version, &wf.IsActive, &wf.InputSchema, &wf.StartStepID); err != nil {
			continue
		}
		list = append(list, wf)
	}
	return list, nil
}

// Get workflow name (for GetExecution response)
func GetWorkflowName(id uuid.UUID) string {
	wf, err := GetWorkflow(context.Background(), id)
	if err != nil {
		return "Unknown"
	}
	return wf.Name
}

// UpdateWorkflowStartStep updates the start step used for a workflow.
func UpdateWorkflowStartStep(ctx context.Context, workflowID uuid.UUID, stepID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE workflows SET start_step_id=$1, updated_at=NOW() WHERE id=$2`,
		stepID, workflowID,
	)
	return err
}
