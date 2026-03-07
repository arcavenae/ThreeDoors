package dispatch

import (
	"context"
	"fmt"
	"time"
)

// GuardrailViolation represents a guardrail check failure with a user-visible message.
type GuardrailViolation struct {
	Message string
}

func (v *GuardrailViolation) Error() string {
	return v.Message
}

// GuardrailChecker validates dispatch preconditions against configured limits.
type GuardrailChecker struct {
	config     DevDispatchConfig
	dispatcher Dispatcher
	audit      *AuditLogger
}

// NewGuardrailChecker creates a GuardrailChecker with the given config, dispatcher, and audit logger.
func NewGuardrailChecker(config DevDispatchConfig, dispatcher Dispatcher, audit *AuditLogger) *GuardrailChecker {
	return &GuardrailChecker{
		config:     config,
		dispatcher: dispatcher,
		audit:      audit,
	}
}

// CheckAll runs all guardrail checks for the given task ID.
// Returns nil if all checks pass, or a GuardrailViolation with a user-visible message.
func (g *GuardrailChecker) CheckAll(ctx context.Context, taskID string) error {
	if err := g.CheckMaxConcurrent(ctx); err != nil {
		return err
	}
	if err := g.CheckCooldown(taskID); err != nil {
		return err
	}
	if err := g.CheckDailyLimit(); err != nil {
		return err
	}
	return nil
}

// CheckMaxConcurrent verifies that the number of active workers is below the configured limit.
func (g *GuardrailChecker) CheckMaxConcurrent(ctx context.Context) error {
	workers, err := g.dispatcher.ListWorkers(ctx)
	if err != nil {
		return fmt.Errorf("check max concurrent: %w", err)
	}

	if len(workers) >= g.config.MaxConcurrent {
		return &GuardrailViolation{
			Message: fmt.Sprintf("Max concurrent workers reached (%d/%d). Wait for a worker to complete.", len(workers), g.config.MaxConcurrent),
		}
	}

	return nil
}

// CheckCooldown verifies that the minimum cooldown period has elapsed since
// the last dispatch of the same task.
func (g *GuardrailChecker) CheckCooldown(taskID string) error {
	if g.audit == nil {
		return nil
	}

	lastDispatch, err := g.audit.LastDispatchForTask(taskID)
	if err != nil {
		return fmt.Errorf("check cooldown: %w", err)
	}

	if lastDispatch.IsZero() {
		return nil
	}

	cooldown := time.Duration(g.config.CooldownMinutes) * time.Minute
	elapsed := time.Since(lastDispatch)
	if elapsed < cooldown {
		remaining := cooldown - elapsed
		return &GuardrailViolation{
			Message: fmt.Sprintf("Cooldown active for task %s. %s remaining before re-dispatch.", taskID, remaining.Truncate(time.Second)),
		}
	}

	return nil
}

// CheckDailyLimit verifies that the daily dispatch count has not been exceeded.
func (g *GuardrailChecker) CheckDailyLimit() error {
	if g.audit == nil {
		return nil
	}

	count, err := g.audit.CountDispatchesToday()
	if err != nil {
		return fmt.Errorf("check daily limit: %w", err)
	}

	if count >= g.config.DailyLimit {
		return &GuardrailViolation{
			Message: fmt.Sprintf("Daily dispatch limit reached (%d/%d).", count, g.config.DailyLimit),
		}
	}

	return nil
}

// BuildDryRunCommand returns the multiclaude command that would be executed
// for the given queue item, without executing it.
func BuildDryRunCommand(item QueueItem) string {
	desc := BuildTaskDescription(item)
	return fmt.Sprintf("multiclaude worker create %q", desc)
}
