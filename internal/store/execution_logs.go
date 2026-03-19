package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ExecutionLog struct {
	ID               uuid.UUID                `json:"id"`
	ExecutionID      uuid.UUID                `json:"execution_id"`
	StepID           *uuid.UUID               `json:"step_id"`
	StepName         string                   `json:"step_name"`
	StepType         string                   `json:"step_type"`
	EvaluatedRules   []map[string]interface{} `json:"evaluated_rules"`
	SelectedNextStep *uuid.UUID               `json:"selected_next_step"`
	Status           string                   `json:"status"` // pending, completed, failed
	ApproverID       *uuid.UUID               `json:"approver_id"`
	ErrorMessage     *string                  `json:"error_message"`
	StartedAt        *time.Time               `json:"started_at"`
	EndedAt          *time.Time               `json:"ended_at"`
}

// Insert execution log
func InsertExecutionLog(ctx context.Context, log *ExecutionLog) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO execution_logs(
			id,execution_id,step_id,step_name,step_type,evaluated_rules,selected_next_step,
			status,approver_id,error_message,started_at,ended_at
		) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		log.ID, log.ExecutionID, log.StepID, log.StepName, log.StepType, log.EvaluatedRules, log.SelectedNextStep,
		log.Status, log.ApproverID, log.ErrorMessage, log.StartedAt, log.EndedAt,
	)
	return err
}

// Update execution log
func UpdateExecutionLog(ctx context.Context, logEntry *ExecutionLog) error {
	_, err := pool.Exec(ctx,
		`UPDATE execution_logs SET step_name=$1,step_type=$2,evaluated_rules=$3,selected_next_step=$4,status=$5,approver_id=$6,error_message=$7,started_at=$8,ended_at=$9 WHERE id=$10`,
		logEntry.StepName, logEntry.StepType, logEntry.EvaluatedRules, logEntry.SelectedNextStep,
		logEntry.Status, logEntry.ApproverID, logEntry.ErrorMessage, logEntry.StartedAt, logEntry.EndedAt, logEntry.ID,
	)
	return err
}

// List execution logs by execution ID
func ListExecutionLogs(ctx context.Context, executionID uuid.UUID) ([]*ExecutionLog, error) {
	rows, err := pool.Query(ctx,
		`SELECT id,execution_id,step_id,step_name,step_type,evaluated_rules,selected_next_step,
		        status,approver_id,error_message,started_at,ended_at
		 FROM execution_logs WHERE execution_id=$1 ORDER BY started_at ASC`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*ExecutionLog
	for rows.Next() {
		log := &ExecutionLog{}
		if err := rows.Scan(&log.ID, &log.ExecutionID, &log.StepID, &log.StepName, &log.StepType,
			&log.EvaluatedRules, &log.SelectedNextStep, &log.Status, &log.ApproverID,
			&log.ErrorMessage, &log.StartedAt, &log.EndedAt); err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}
