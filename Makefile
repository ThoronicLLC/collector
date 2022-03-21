# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTOOL=$(GOCMD) tool
GOTEST=$(GOCMD) test
GOTESTRACE=$(GOTEST) -race
GOGET=$(GOCMD) get
GOFMT=$(GOCMD)fmt

all: ## Build package in all supported formats
	$(CURDIR)/build/package/build.bash --osx --linux --windows --clean --verbose

all_zip: ## Build package in all supported formats and compress
	$(CURDIR)/build/package/build.bash --osx --linux --windows --clean --verbose --zip

linux: ## Build for linux
	$(CURDIR)/build/package/build.bash --linux --clean --verbose

linux_zip: ## Build for linux and compress it in a zip file
	$(CURDIR)/build/package/build.bash --linux --clean --verbose --zip

osx: ## Build for osx
	$(CURDIR)/build/package/build.bash --osx --clean --verbose

osx_zip: ## Build for osx and compress it in a zip file
	$(CURDIR)/build/package/build.bash --osx --clean --verbose --zip

windows: ## Build for windows
	$(CURDIR)/build/package/build.bash --windows --clean --verbose

windows_zip: ## Build for windows and compress it in a zip file
	$(CURDIR)/build/package/build.bash --windows --clean --verbose --zip

docker: ## Build docker container
	$(CURDIR)/build/docker/build.bash

test: ## Run tests for the project
	$(GOTEST) -count=1 -coverprofile=$(CURDIR)/bin/cover.out -short -cover -failfast ./...

race: ## Run tests for the project (while detecting race conditions)
	$(GOTESTRACE) -coverprofile=$(CURDIR)/bin/cover.out -short -cover -failfast ./...

cov: test ## Run tests with HTML for the project
	$(GOTOOL) cover -html=$(CURDIR)/bin/cover.out

gofmt: ## gofmt code formatting
	@echo Running go formatting with the following command:
	$(GOFMT) -e -s -w .
