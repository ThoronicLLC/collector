# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTOOL=$(GOCMD) tool
GOTEST=$(GOCMD) test
GOTESTRACE=$(GOTEST) -race
GOGET=$(GOCMD) get
GOFMT=$(GOCMD)fmt

dev: ## Build development binaries
	$(CURDIR)/build/package/dev.bash

prod: ## Build production binaries
	$(CURDIR)/build/package/prod.bash

test: ## Run tests for the project
	$(GOTEST) -count=1 -coverprofile=$(CURDIR)/bin/cover.out -short -cover -failfast ./... -v

race: ## Run tests for the project (while detecting race conditions)
	$(GOTESTRACE) -coverprofile=$(CURDIR)/bin/cover.out -short -cover -failfast ./...

cov: test ## Run tests with HTML for the project
	$(GOTOOL) cover -html=$(CURDIR)/bin/cover.out

gofmt: ## gofmt code formatting
	@echo Running go formatting with the following command:
	$(GOFMT) -e -s -w .
