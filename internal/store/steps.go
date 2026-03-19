package store

import (
	"context"

	"github.com/google/uuid"
)

type Step struct {
	ID         uuid.UUID              `json:"id"`
	WorkflowID uuid.UUID              `json:"workflow_id"`
	Name       string                 `json:"name"`
	StepType   string                 `json:"step_type"`
	Order      int                    `json:"order"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Insert step
func InsertStep(ctx context.Context, s *Step) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO steps(id,workflow_id,name,step_type,"order",metadata,created_at,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
		s.ID, s.WorkflowID, s.Name, s.StepType, s.Order, s.Metadata,
	)
	return err
}

// Get step by ID
func GetStepByID(ctx context.Context, id uuid.UUID) (*Step, error) {
	row := pool.QueryRow(ctx,
		`SELECT id,workflow_id,name,step_type,"order",metadata FROM steps WHERE id=$1`, id,
	)
	s := &Step{}
	err := row.Scan(&s.ID, &s.WorkflowID, &s.Name, &s.StepType, &s.Order, &s.Metadata)
	return s, err
}

// List steps by workflow
func ListSteps(ctx context.Context, workflowID uuid.UUID) ([]*Step, error) {
	rows, err := pool.Query(ctx, `SELECT id,workflow_id,name,step_type,"order",metadata FROM steps WHERE workflow_id=$1 ORDER BY "order" ASC`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Step
	for rows.Next() {
		s := &Step{}
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.Name, &s.StepType, &s.Order, &s.Metadata); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}
