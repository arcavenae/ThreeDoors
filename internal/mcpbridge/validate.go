package mcpbridge

import (
	"fmt"
	"regexp"
	"strings"
)

// Validation constants for write tool parameters.
const (
	// MaxBodyLength is the maximum length for message body and task descriptions.
	MaxBodyLength = 2000

	// MaxTaskLength is the maximum length for worker task descriptions.
	MaxTaskLength = 500
)

// Validation patterns for parameter formats.
var (
	recipientPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	messageIDPattern = regexp.MustCompile(`^msg-[a-f0-9-]+$`)
)

// shellMetachars contains characters that could be used for shell injection.
// Defense-in-depth: exec.Command doesn't use a shell, but we reject these
// anyway to prevent exploitation if the execution model ever changes.
var shellMetachars = []string{";", "|", "$(", "`", "&&", "||", ">", "<", "\x00"}

// ValidationError represents an input validation failure with a descriptive message.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// validateNoShellMetachars checks that a string does not contain shell metacharacters.
func validateNoShellMetachars(field, value string) error {
	for _, meta := range shellMetachars {
		if strings.Contains(value, meta) {
			return &ValidationError{
				Field:   field,
				Message: fmt.Sprintf("contains prohibited character sequence %q", meta),
			}
		}
	}
	return nil
}

// validateRecipient checks that a recipient name matches the allowed pattern
// (alphanumeric, hyphens, underscores only).
func validateRecipient(recipient string) error {
	if recipient == "" {
		return &ValidationError{Field: "recipient", Message: "is required"}
	}
	if err := validateNoShellMetachars("recipient", recipient); err != nil {
		return err
	}
	if !recipientPattern.MatchString(recipient) {
		return &ValidationError{
			Field:   "recipient",
			Message: "must contain only alphanumeric characters, hyphens, and underscores",
		}
	}
	return nil
}

// validateMessageID checks that a message ID matches the expected format (msg-<uuid>).
func validateMessageID(id string) error {
	if id == "" {
		return &ValidationError{Field: "message_id", Message: "is required"}
	}
	if err := validateNoShellMetachars("message_id", id); err != nil {
		return err
	}
	if !messageIDPattern.MatchString(id) {
		return &ValidationError{
			Field:   "message_id",
			Message: "must match format msg-<uuid> (e.g., msg-abc123-def4-5678)",
		}
	}
	return nil
}

// validateBody checks that a message body is not empty, not too long, and
// contains no shell metacharacters.
func validateBody(body string) error {
	if body == "" {
		return &ValidationError{Field: "body", Message: "is required"}
	}
	if len(body) > MaxBodyLength {
		return &ValidationError{
			Field:   "body",
			Message: fmt.Sprintf("exceeds maximum length of %d characters (got %d)", MaxBodyLength, len(body)),
		}
	}
	return validateNoShellMetachars("body", body)
}

// validateTask checks that a task description is not empty, within length limits,
// and contains no shell metacharacters.
func validateTask(task string) error {
	if task == "" {
		return &ValidationError{Field: "task", Message: "is required"}
	}
	if len(task) > MaxTaskLength {
		return &ValidationError{
			Field:   "task",
			Message: fmt.Sprintf("exceeds maximum length of %d characters (got %d)", MaxTaskLength, len(task)),
		}
	}
	return validateNoShellMetachars("task", task)
}
