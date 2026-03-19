package engine

import (
	"context"
	"log"
	"time"

	"halleyx-workflow-docker/internal/store"
	"halleyx-workflow-docker/internal/ws"

	"github.com/google/uuid"
)

func RunExecution(execID uuid.UUID) {
	ctx := context.Background()

	exec, err := store.GetExecution(ctx, execID)
	if err != nil {
		log.Println("Execution not found:", err)
		return
	}

	// Mark execution as in_progress and set start time
	if exec.Status != "in_progress" {
		exec.Status = "in_progress"
		now := time.Now()
		exec.StartedAt = &now
		store.UpdateExecution(ctx, exec)
	}

	workflow, err := store.GetWorkflow(ctx, exec.WorkflowID)
	if err != nil {
		log.Println("Workflow not found:", err)
		store.MarkExecutionFailed(ctx, execID, "workflow not found")
		return
	}

	// Runtime context
	runtimeCtx := make(map[string]interface{})
	for k, v := range exec.Data {
		runtimeCtx[k] = v
	}

	// Determine initial current step
	currentStepID := uuid.Nil
	if exec.CurrentStepID != nil {
		currentStepID = *exec.CurrentStepID
	} else if workflow.StartStepID != nil {
		currentStepID = *workflow.StartStepID
	} else {
		// If no explicit start step, pick the first step by order
		steps, err := store.ListSteps(ctx, workflow.ID)
		if err == nil && len(steps) > 0 {
			first := steps[0]
			for _, s := range steps {
				if s.Order < first.Order {
					first = s
				}
			}
			currentStepID = first.ID
		}
	}
	visitCount := make(map[uuid.UUID]int)
	const maxVisits = 50

	for currentStepID != uuid.Nil {
		// Check for cancellation
		currentExec, _ := store.GetExecution(ctx, execID)
		if currentExec.Status == "canceled" {
			return
		}

		visitCount[currentStepID]++
		if visitCount[currentStepID] > maxVisits {
			store.MarkExecutionFailed(ctx, execID, "max iterations exceeded")
			return
		}

		step, err := store.GetStepByID(ctx, currentStepID)
		if err != nil {
			log.Println("Step not found:", err)
			store.MarkExecutionFailed(ctx, execID, "step not found")
			return
		}

		// Update current step in execution record
		store.UpdateExecutionCurrentStep(ctx, execID, &step.ID)

		// Start step log
		logEntry := &store.ExecutionLog{
			ID:          uuid.New(),
			ExecutionID: execID,
			StepID:      &step.ID,
			StepName:    step.Name,
			StepType:    step.StepType,
			Status:      "pending",
			StartedAt:   ptrTime(time.Now()),
		}
		if err := store.InsertExecutionLog(ctx, logEntry); err != nil {
			log.Println("Failed to insert execution log:", err)
		}

		// ---- Broadcast STEP_STARTED ----
		ws.BroadcastEvent("STEP_STARTED", map[string]interface{}{
			"execution_id": execID.String(),
			"step_id":      step.ID.String(),
			"status":       "in_progress",
		})

		// Execute step
		outputs, execErr := ExecuteStep(step, runtimeCtx, execID)
		if execErr != nil {
			logEntry.Status = "failed"
			logEntry.EndedAt = ptrTime(time.Now())
			errMsg := execErr.Error()
			logEntry.ErrorMessage = &errMsg
			store.UpdateExecutionLog(ctx, logEntry)
			store.MarkExecutionFailed(ctx, execID, execErr.Error())

			log.Printf("Execution %s step %s failed: %v", execID, step.Name, execErr)

			// ---- Broadcast STEP_FAILED ----
			payload := map[string]interface{}{
				"execution_id": execID.String(),
				"step_id":      step.ID.String(),
				"status":       "failed",
				"error":        execErr.Error(),
			}
			// Include any logs/output if available
			if len(outputs) > 0 {
				payload["logs"] = outputs
			}
			ws.BroadcastEvent("STEP_FAILED", payload)
			return
		}

		// Merge outputs into runtime context
		for k, v := range outputs {
			runtimeCtx[k] = v
		}

		// Persist runtime context back to execution
		store.UpdateExecutionData(ctx, execID, runtimeCtx)

		// If this step is an approval step, pause execution until external approval
		if pending, ok := outputs["approval_pending"].(bool); ok && pending {
			exec.Status = "pending"
			logEntry.Status = "pending"
			logEntry.EndedAt = ptrTime(time.Now())
			store.UpdateExecutionLog(ctx, logEntry)
			store.UpdateExecution(ctx, exec) // ensure status is persisted
			ws.BroadcastEvent("STEP_PENDING", map[string]interface{}{
				"execution_id": execID.String(),
				"step_id":      step.ID.String(),
				"status":       "pending",
			})
			return
		}

		// Evaluate rules for next step
		nextStep, evaluatedRules, evalErr := EvaluateRulesForStep(ctx, step.ID, runtimeCtx)
		logEntry.EvaluatedRules = evaluatedRules
		if nextStep != nil {
			logEntry.SelectedNextStep = &nextStep.ID
		}
		if evalErr != nil {
			// store but continue (treat as no match)
			log.Println("Rule evaluation error:", evalErr)
		}

		// Mark step completed
		logEntry.Status = "completed"
		logEntry.EndedAt = ptrTime(time.Now())
		store.UpdateExecutionLog(ctx, logEntry)

		// ---- Broadcast STEP_COMPLETED ----
		ws.BroadcastEvent("STEP_COMPLETED", map[string]interface{}{
			"execution_id": execID.String(),
			"step_id":      step.ID.String(),
			"status":       "completed",
			"logs":         outputs,
		})

		if nextStep == nil {
			store.MarkExecutionCompleted(ctx, execID)
			ws.BroadcastEvent("EXECUTION_COMPLETED", map[string]interface{}{
				"execution_id": execID.String(),
				"status":       "completed",
			})
			return
		}

		// Move to next step
		currentStepID = nextStep.ID
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
