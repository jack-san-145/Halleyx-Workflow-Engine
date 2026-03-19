package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Execution struct {
	ID              uuid.UUID              `json:"id"`
	WorkflowID      uuid.UUID              `json:"workflow_id"`
	WorkflowVersion int                    `json:"workflow_version"`
	Status          string                 `json:"status"` // pending, in_progress, completed, failed, canceled
	Data            map[string]interface{} `json:"data"`
	CurrentStepID   *uuid.UUID             `json:"current_step_id"`
	Retries         int                    `json:"retries"`
	TriggeredBy     *uuid.UUID             `json:"triggered_by"`
	StartedAt       *time.Time             `json:"started_at"`
	EndedAt         *time.Time             `json:"ended_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Insert a new execution
func InsertExecution(ctx context.Context, e *Execution) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO executions(id,workflow_id,workflow_version,status,data,current_step_id,retries,triggered_by,started_at,ended_at,created_at,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())`,
		e.ID, e.WorkflowID, e.WorkflowVersion, e.Status, e.Data, e.CurrentStepID, e.Retries, e.TriggeredBy, e.StartedAt, e.EndedAt,
	)
	return err
}

// Get execution by ID
func GetExecution(ctx context.Context, id uuid.UUID) (*Execution, error) {
	row := pool.QueryRow(ctx,
		`SELECT id,workflow_id,workflow_version,status,data,current_step_id,retries,triggered_by,started_at,ended_at,created_at,updated_at
		 FROM executions WHERE id=$1`, id,
	)
	e := &Execution{}
	err := row.Scan(&e.ID, &e.WorkflowID, &e.WorkflowVersion, &e.Status, &e.Data, &e.CurrentStepID, &e.Retries, &e.TriggeredBy, &e.StartedAt, &e.EndedAt, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

// Update execution (status, retries, current_step_id, data, timestamps, etc.)
func UpdateExecution(ctx context.Context, e *Execution) error {
	_, err := pool.Exec(ctx,
		`UPDATE executions SET status=$1,current_step_id=$2,retries=$3,started_at=$4,ended_at=$5,data=$6,updated_at=NOW() WHERE id=$7`,
		e.Status, e.CurrentStepID, e.Retries, e.StartedAt, e.EndedAt, e.Data, e.ID,
	)
	return err
}

// UpdateExecutionCurrentStep updates only the current step of an execution.
func UpdateExecutionCurrentStep(ctx context.Context, executionID uuid.UUID, stepID *uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE executions SET current_step_id=$1,updated_at=NOW() WHERE id=$2`,
		stepID, executionID,
	)
	return err
}

// UpdateExecutionData updates only the execution context data.
func UpdateExecutionData(ctx context.Context, executionID uuid.UUID, data map[string]interface{}) error {
	_, err := pool.Exec(ctx,
		`UPDATE executions SET data=$1,updated_at=NOW() WHERE id=$2`,
		data, executionID,
	)
	return err
}

// Mark execution as completed
func MarkExecutionCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := pool.Exec(ctx,
		`UPDATE executions SET status='completed',ended_at=$1,updated_at=NOW() WHERE id=$2`, now, id)
	return err
}

// Mark execution as failed
func MarkExecutionFailed(ctx context.Context, id uuid.UUID, msg string) error {
	now := time.Now()
	_, err := pool.Exec(ctx,
		`UPDATE executions SET status='failed',ended_at=$1,updated_at=NOW() WHERE id=$2`, now, id)
	return err
}

// Mark execution as canceled
func MarkExecutionCanceled(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := pool.Exec(ctx,
		`UPDATE executions SET status='canceled',ended_at=$1,updated_at=NOW() WHERE id=$2`, now, id)
	return err
}
