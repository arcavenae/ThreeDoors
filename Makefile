THREEDOORS_DIR ?= $(HOME)/.threedoors
VERSION ?= dev

.PHONY: build run clean fmt lint test test-docker analyze test-scripts sign pkg release-local test-dist

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/threedoors ./cmd/threedoors

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

test-docker:
	@command -v docker >/dev/null 2>&1 || { echo "Error: Docker is required but not found. Install from https://docs.docker.com/get-docker/"; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "Error: Docker daemon is not running. Start Docker and try again."; exit 1; }
	@mkdir -p test-results
	docker compose -f docker-compose.test.yml run --rm test

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

sign:
ifndef APPLE_SIGNING_IDENTITY
	@echo "APPLE_SIGNING_IDENTITY not set, skipping signing"
else
	codesign --force --options runtime --sign "$(APPLE_SIGNING_IDENTITY)" --timestamp bin/threedoors
endif

pkg:
ifndef APPLE_INSTALLER_IDENTITY
	@echo "APPLE_INSTALLER_IDENTITY not set, skipping pkg creation"
else
	@chmod +x scripts/create-pkg.sh
	./scripts/create-pkg.sh bin/threedoors "$(VERSION)" "$(APPLE_INSTALLER_IDENTITY)" bin/threedoors.pkg
endif

release-local: build sign pkg

test-dist: build
	@echo "=== Distribution Tests ==="
	@echo "Testing --version flag..."
	@./bin/threedoors --version | grep -q "ThreeDoors" && echo "  PASS" || (echo "  FAIL" && exit 1)
	@echo "Testing Homebrew formula syntax..."
	@ruby -c Formula/threedoors.rb > /dev/null 2>&1 && echo "  PASS" || (echo "  FAIL" && exit 1)
	@echo "Testing shell script syntax..."
	@bash -n scripts/create-pkg.sh && echo "  PASS" || (echo "  FAIL" && exit 1)
	@echo "Testing make sign dry-run..."
	@make -n sign > /dev/null 2>&1 && echo "  PASS" || (echo "  FAIL" && exit 1)
	@echo "Testing make pkg dry-run..."
	@make -n pkg > /dev/null 2>&1 && echo "  PASS" || (echo "  FAIL" && exit 1)
	@echo "All distribution tests passed."
