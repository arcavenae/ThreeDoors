//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Task represents a parsed task from Apple Notes.
type Task struct {
	Text   string `json:"text"`
	Status string `json:"status"`
}

// BenchmarkResult holds latency measurements.
type BenchmarkResult struct {
	Approach      string  `json:"approach"`
	Operation     string  `json:"operation"`
	Iterations    int     `json:"iterations"`
	P50Ms         float64 `json:"p50_ms"`
	P95Ms         float64 `json:"p95_ms"`
	MaxMs         float64 `json:"max_ms"`
	NotesAppState string  `json:"notes_app_state"`
}

func runOsascript(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func readNote(noteTitle string) ([]Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	script := fmt.Sprintf(`tell application "Notes" to get plaintext text of note "%s"`, noteTitle)
	output, err := runOsascript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("osascript read failed: %w", err)
	}

	return parseTasks(output), nil
}

func parseTasks(body string) []Task {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	var tasks []Task
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		text, status := parseCheckbox(trimmed)
		tasks = append(tasks, Task{Text: text, Status: status})
	}
	return tasks
}

func parseCheckbox(line string) (string, string) {
	prefixes := []struct {
		prefix string
		status string
	}{
		{"- [x] ", "complete"},
		{"- [X] ", "complete"},
		{"* [x] ", "complete"},
		{"* [X] ", "complete"},
		{"- [ ] ", "todo"},
		{"* [ ] ", "todo"},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(line, p.prefix) {
			return strings.TrimSpace(line[len(p.prefix):]), p.status
		}
	}
	return line, "todo"
}

func benchmarkRead(noteTitle string, iterations int) (*BenchmarkResult, error) {
	var durations []float64

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := readNote(noteTitle)
		elapsed := time.Since(start).Seconds() * 1000 // ms
		if err != nil {
			return nil, fmt.Errorf("benchmark iteration %d failed: %w", i, err)
		}
		durations = append(durations, elapsed)
	}

	sort.Float64s(durations)

	return &BenchmarkResult{
		Approach:      "applescript",
		Operation:     "read",
		Iterations:    iterations,
		P50Ms:         math.Round(percentile(durations, 0.50)*100) / 100,
		P95Ms:         math.Round(percentile(durations, 0.95)*100) / 100,
		MaxMs:         math.Round(durations[len(durations)-1]*100) / 100,
		NotesAppState: "unknown",
	}, nil
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper {
		return sorted[lower]
	}
	weight := idx - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func main() {
	noteTitle := "ThreeDoors Tasks"

	fmt.Println("=== Go AppleScript PoC ===")
	fmt.Printf("Reading note: %s\n", noteTitle)
	fmt.Println("---")

	tasks, err := readNote(noteTitle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] %v\n", err)
		os.Exit(1)
	}

	for i, t := range tasks {
		statusLabel := "TODO"
		if t.Status == "complete" {
			statusLabel = "DONE"
		}
		fmt.Printf("%d. [%s] %s\n", i+1, statusLabel, t.Text)
	}
	fmt.Println("---")
	fmt.Printf("[OK] Read %d tasks from %q\n\n", len(tasks), noteTitle)

	// Benchmark
	fmt.Println("=== Benchmarking (10 iterations) ===")
	result, err := benchmarkRead(noteTitle, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FAIL] Benchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("p50: %.2f ms\n", result.P50Ms)
	fmt.Printf("p95: %.2f ms\n", result.P95Ms)
	fmt.Printf("max: %.2f ms\n", result.MaxMs)

	// Write benchmark JSON
	benchJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\nBenchmark JSON:\n%s\n", string(benchJSON))

	// Save to benchmarks dir
	if err := os.MkdirAll("benchmarks", 0o755); err == nil {
		if err := os.WriteFile("benchmarks/applescript_read.json", benchJSON, 0o644); err == nil {
			fmt.Println("\n[OK] Benchmark saved to benchmarks/applescript_read.json")
		}
	}
}
