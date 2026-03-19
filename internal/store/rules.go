package store

import (
	"context"

	"github.com/google/uuid"
)

type Rule struct {
	ID         uuid.UUID  `json:"id"`
	StepID     uuid.UUID  `json:"step_id"`
	Condition  string     `json:"condition"`
	NextStepID *uuid.UUID `json:"next_step_id"`
	Priority   int        `json:"priority"`
}

// Insert rule
func InsertRule(ctx context.Context, r *Rule) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO rules(id,step_id,condition,next_step_id,priority,created_at,updated_at)
		 VALUES($1,$2,$3,$4,$5,NOW(),NOW())`,
		r.ID, r.StepID, r.Condition, r.NextStepID, r.Priority,
	)
	return err
}

// List rules for step
func ListRules(ctx context.Context, stepID uuid.UUID) ([]*Rule, error) {
	rows, err := pool.Query(ctx,
		`SELECT id,step_id,condition,next_step_id,priority FROM rules WHERE step_id=$1 ORDER BY priority ASC`, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Rule
	for rows.Next() {
		r := &Rule{}
		if err := rows.Scan(&r.ID, &r.StepID, &r.Condition, &r.NextStepID, &r.Priority); err != nil {
			continue
		}
		list = append(list, r)
	}
	return list, nil
}

// GetRulesForStep returns the rules for a step (ordered by priority).
func GetRulesForStep(ctx context.Context, stepID uuid.UUID) ([]*Rule, error) {
	return ListRules(ctx, stepID)
}
