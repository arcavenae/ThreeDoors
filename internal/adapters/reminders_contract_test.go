package adapters_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/adapters/reminders"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// statefulExecutor simulates Apple Reminders state for contract tests.
// It dispatches based on JXA script content patterns.
type statefulExecutor struct {
	mu     sync.Mutex
	store  map[string]reminders.ReminderJSON
	nextID int
}

func newStatefulExecutor() *statefulExecutor {
	return &statefulExecutor{
		store: make(map[string]reminders.ReminderJSON),
	}
}

var (
	idPattern       = regexp.MustCompile(`=== "([^"]+)"`)
	namePattern     = regexp.MustCompile(`name: "([^"]*)"`)
	bodyPattern     = regexp.MustCompile(`body: "([^"]*)"`)
	priorityPattern = regexp.MustCompile(`priority: (\d+)`)
	updateName      = regexp.MustCompile(`\.name = "([^"]*)"`)
	updateBody      = regexp.MustCompile(`\.body = "([^"]*)"`)
	updatePriority  = regexp.MustCompile(`\.priority = (\d+)`)
)

func (e *statefulExecutor) Execute(_ context.Context, script string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch {
	case strings.Contains(script, "names.push"):
		// ReadLists
		data, _ := json.Marshal([]string{"Test"})
		return string(data), nil

	case strings.Contains(script, "completed: false"):
		// ReadReminders — return all non-completed reminders
		var result []reminders.ReminderJSON
		for _, r := range e.store {
			if !r.Completed {
				result = append(result, r)
			}
		}
		if result == nil {
			result = []reminders.ReminderJSON{}
		}
		data, _ := json.Marshal(result)
		return string(data), nil

	case strings.Contains(script, "completed = true"):
		// CompleteReminder
		id := extractMatch(idPattern, script)
		if r, ok := e.store[id]; ok {
			r.Completed = true
			r.ModificationDate = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
			e.store[id] = r
			return `{"success":true}`, nil
		}
		return `{"success":false,"error":"reminder not found"}`, nil

	case strings.Contains(script, "app.Reminder({"):
		// CreateReminder
		e.nextID++
		id := fmt.Sprintf("x-apple-reminder://test-%d", e.nextID)
		now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		name := extractMatch(namePattern, script)
		body := extractMatch(bodyPattern, script)
		priority := extractIntMatch(priorityPattern, script)
		e.store[id] = reminders.ReminderJSON{
			ID:               id,
			Name:             name,
			Body:             body,
			Priority:         priority,
			CreationDate:     now,
			ModificationDate: now,
		}
		return fmt.Sprintf(`{"success":true,"id":"%s"}`, id), nil

	case strings.Contains(script, ".name = "):
		// UpdateReminder
		id := extractMatch(idPattern, script)
		if r, ok := e.store[id]; ok {
			r.Name = extractMatch(updateName, script)
			r.Body = extractMatch(updateBody, script)
			r.Priority = extractIntMatch(updatePriority, script)
			r.ModificationDate = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
			e.store[id] = r
			return `{"success":true}`, nil
		}
		return `{"success":false,"error":"reminder not found"}`, nil

	case strings.Contains(script, "app.delete"):
		// DeleteReminder
		id := extractMatch(idPattern, script)
		if _, ok := e.store[id]; ok {
			delete(e.store, id)
			return `{"success":true}`, nil
		}
		return `{"success":false,"error":"reminder not found"}`, nil
	}

	return `[]`, nil
}

func extractMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func extractIntMatch(re *regexp.Regexp, s string) int {
	s2 := extractMatch(re, s)
	n, _ := strconv.Atoi(s2)
	return n
}

func TestRemindersProviderContract(t *testing.T) {
	t.Parallel()

	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		return reminders.NewRemindersProvider(newStatefulExecutor(), []string{"Test"})
	}

	adapters.RunContractTests(t, factory)
}
