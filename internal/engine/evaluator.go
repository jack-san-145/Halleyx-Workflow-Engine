package engine

import (
	"context"
	"fmt"
	"halleyx-workflow-docker/internal/store"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
)

func EvaluateCondition(cond string, ctx map[string]interface{}) (bool, error) {
	if strings.TrimSpace(strings.ToUpper(cond)) == "DEFAULT" {
		return true, nil
	}
	expr, err := govaluate.NewEvaluableExpression(cond)
	if err != nil {
		return false, err
	}
	res, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	b, ok := res.(bool)
	if !ok {
		return false, fmt.Errorf("condition did not return boolean")
	}
	return b, nil
}

// EvaluateRulesForStep evaluates all rules for a step and returns the selected next step (nil if none).
// It also returns the list of evaluated rules for logging.
func EvaluateRulesForStep(ctx context.Context, stepID uuid.UUID, runtimeCtx map[string]interface{}) (*store.Step, []map[string]interface{}, error) {
	rules, err := store.GetRulesForStep(ctx, stepID)
	if err != nil {
		return nil, nil, err
	}

	evaluated := []map[string]interface{}{}
	for _, r := range rules {
		ok, evalErr := EvaluateCondition(r.Condition, runtimeCtx)
		evaluated = append(evaluated, map[string]interface{}{
			"rule":   r.Condition,
			"result": ok,
		})

		if evalErr != nil {
			continue
		}

		if ok {
			if r.NextStepID == nil {
				return nil, evaluated, nil
			}
			nextStep, err := store.GetStepByID(ctx, *r.NextStepID)
			if err != nil {
				return nil, evaluated, err
			}
			return nextStep, evaluated, nil
		}
	}

	return nil, evaluated, nil
}
