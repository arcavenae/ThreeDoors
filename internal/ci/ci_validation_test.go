package ci_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// projectRoot returns the repository root by walking up from the test file location.
func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	// internal/ci/ci_validation_test.go -> repo root is two directories up
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func TestToolVersionConsistency(t *testing.T) {
	root := projectRoot(t)

	tests := []struct {
		name       string
		docFile    string
		configFile string
		docPattern *regexp.Regexp
		cfgPattern *regexp.Regexp
		compareFn  func(docVer, cfgVer string) bool
	}{
		{
			name:       "golangci-lint doc version matches config major version",
			docFile:    "docs/architecture/tech-stack.md",
			configFile: ".golangci.yml",
			docPattern: regexp.MustCompile(`golangci-lint\s*\|\s*(\d+)\.\d+\.\d+`),
			cfgPattern: regexp.MustCompile(`version:\s*"(\d+)"`),
			compareFn: func(docVer, cfgVer string) bool {
				return docVer == cfgVer
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docContent, err := os.ReadFile(filepath.Join(root, tt.docFile))
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.docFile, err)
			}

			cfgContent, err := os.ReadFile(filepath.Join(root, tt.configFile))
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.configFile, err)
			}

			docMatches := tt.docPattern.FindSubmatch(docContent)
			if docMatches == nil {
				t.Fatalf("version pattern not found in %s", tt.docFile)
				return
			}

			cfgMatches := tt.cfgPattern.FindSubmatch(cfgContent)
			if cfgMatches == nil {
				t.Fatalf("version pattern not found in %s", tt.configFile)
				return
			}

			docVer := strings.TrimSpace(string(docMatches[1]))
			cfgVer := strings.TrimSpace(string(cfgMatches[1]))

			if !tt.compareFn(docVer, cfgVer) {
				t.Errorf("version mismatch: %s has major version %s, but %s has version %s",
					tt.docFile, docVer, tt.configFile, cfgVer)
			}
		})
	}
}
