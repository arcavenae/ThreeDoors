package retrospector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadFindings reads all findings from a JSONL file at the given path.
func ReadFindings(path string) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open findings %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return readFindingsFrom(f)
}

func readFindingsFrom(r io.Reader) ([]Finding, error) {
	var findings []Finding
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var f Finding
		if err := json.Unmarshal(line, &f); err != nil {
			return nil, fmt.Errorf("parse finding line %d: %w", lineNum, err)
		}
		findings = append(findings, f)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan findings: %w", err)
	}
	return findings, nil
}
