.PHONY: help setup install backend-deps frontend-deps check-env require-go require-npm up open down logs restart dev backend-run frontend-run test build clean

COMPOSE ?= docker compose
BACKEND_DIR := backend
FRONTEND_DIR := frontend

help:
	@echo "Krystal Network Tools - available targets"
	@echo ""
	@echo "Setup"
	@echo "  make setup         Create .env if missing and install dependencies"
	@echo "  make install       Install backend and frontend dependencies"
	@echo ""
	@echo "Run (recommended)"
	@echo "  make up            Start full stack with Docker Compose (PORT=... optional)"
	@echo "  make open          Open app URL in browser when healthy"
	@echo "  make down          Stop Docker Compose services"
	@echo "  make logs          Tail Docker Compose logs"
	@echo "  make restart       Restart Docker Compose services"
	@echo ""
	@echo "Run (without Docker)"
	@echo "  make dev           Run backend and frontend in separate terminals"
	@echo "  make backend-run   Run backend locally"
	@echo "  make frontend-run  Run frontend locally"
	@echo ""
	@echo "Quality/build"
	@echo "  make test          Run backend and frontend tests"
	@echo "  make build         Build backend binary and frontend bundle"
	@echo "  make clean         Remove common generated artifacts"

setup: check-env install

install: backend-deps frontend-deps

backend-deps:
	@if command -v go >/dev/null 2>&1; then \
		echo "Downloading Go modules..."; \
		go -C $(BACKEND_DIR) mod download; \
	else \
		echo "Skipping backend deps (go not found)."; \
		echo "Install Go for local backend development, or use 'make up' for Docker."; \
	fi

frontend-deps:
	@if command -v npm >/dev/null 2>&1; then \
		echo "Installing frontend npm dependencies..."; \
		npm --prefix $(FRONTEND_DIR) ci; \
	else \
		echo "Skipping frontend deps (npm not found)."; \
		echo "Install Node.js/npm for local frontend development, or use 'make up' for Docker."; \
	fi

check-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from defaults..."; \
		port=8080; \
		while lsof -nP -iTCP:$$port -sTCP:LISTEN >/dev/null 2>&1; do \
			port=$$((port + 1)); \
		done; \
		printf "DNS_SERVER=1.1.1.1\nHTTPS_HOST=localhost\nPORT=%s\n" "$$port" > .env; \
		echo "Selected PORT=$$port"; \
	else \
		echo ".env already exists"; \
	fi

up:
	@$(COMPOSE) up --build -d
	@port="$${PORT:-$$(awk -F= '/^PORT=/{print $$2}' .env 2>/dev/null)}"; \
	if [ -z "$$port" ]; then port=8080; fi; \
	echo "Application should be available on http://localhost:$$port"

open:
	@port="$${PORT:-$$(awk -F= '/^PORT=/{print $$2}' .env 2>/dev/null)}"; \
	if [ -z "$$port" ]; then port=8080; fi; \
	url="http://localhost:$$port"; \
	echo "Waiting for $$url ..."; \
	for i in $$(seq 1 30); do \
		if curl -fsS "$$url" >/dev/null 2>&1; then \
			echo "Opening $$url"; \
			if command -v open >/dev/null 2>&1; then \
				open "$$url"; \
			elif command -v xdg-open >/dev/null 2>&1; then \
				xdg-open "$$url" >/dev/null 2>&1; \
			else \
				echo "No browser opener found. Open manually: $$url"; \
			fi; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "Service not responding yet at $$url. Try: make logs"

down:
	@$(COMPOSE) down

logs:
	@$(COMPOSE) logs -f

restart: down up

dev:
	@echo "Run these in separate terminals:"
	@echo "  make backend-run"
	@echo "  make frontend-run"

require-go:
	@command -v go >/dev/null 2>&1 || (echo "go is required for this target. Use 'make up' for Docker-based run." && exit 1)

require-npm:
	@command -v npm >/dev/null 2>&1 || (echo "npm is required for this target. Use 'make up' for Docker-based run." && exit 1)

backend-run: require-go
	@go -C $(BACKEND_DIR) run . http

frontend-run: require-npm
	@npm --prefix $(FRONTEND_DIR) start

test: require-go require-npm
	@go -C $(BACKEND_DIR) test ./...
	@npm --prefix $(FRONTEND_DIR) test -- --watchAll=false

build: require-go require-npm
	@go -C $(BACKEND_DIR) build -o main
	@npm --prefix $(FRONTEND_DIR) run build

clean:
	@rm -f $(BACKEND_DIR)/main
	@rm -rf $(FRONTEND_DIR)/build
