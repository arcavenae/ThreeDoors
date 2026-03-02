THREEDOORS_DIR ?= $(HOME)/.threedoors

.PHONY: build run clean fmt lint test analyze test-scripts

build:
	go build -o bin/threedoors ./cmd/threedoors

run: build
	./bin/threedoors

clean:
	rm -rf bin/

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

test:
	go test ./... -v

analyze:
	@chmod +x scripts/*.sh
	@echo "=== Session Analysis ==="
	@./scripts/analyze_sessions.sh $(THREEDOORS_DIR)/sessions.jsonl
	@echo ""
	@echo "=== Daily Completions ==="
	@./scripts/daily_completions.sh $(THREEDOORS_DIR)/completed.txt
	@echo ""
	@echo "=== Validation Decision ==="
	@./scripts/validation_decision.sh $(THREEDOORS_DIR)/sessions.jsonl

test-scripts:
	@chmod +x scripts/*.sh
	@echo "Testing analyze_sessions.sh..."
	@./scripts/analyze_sessions.sh scripts/testdata/sessions.jsonl > /dev/null
	@echo "  PASS"
	@echo "Testing daily_completions.sh..."
	@./scripts/daily_completions.sh scripts/testdata/completed.txt > /dev/null
	@echo "  PASS"
	@echo "Testing validation_decision.sh..."
	@./scripts/validation_decision.sh scripts/testdata/sessions.jsonl > /dev/null
	@echo "  PASS"
	@echo "All script tests passed."
