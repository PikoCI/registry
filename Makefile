GOCACHE := $(shell go env GOCACHE)

.PHONY: help
help: Makefile ## This help dialog
	@IFS=$$'\n' ; \
	help_lines=(`grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

export GOCACHE

-include .env
export

.PHONY: serve
serve: ## Starts the dev server with in-memory SQLite
	@go run . server -p 8080 --log-level=debug --jwt-secret=dev-secret --db-system=sqlite --github-client-id=$(GITHUB_CLIENT_ID) --github-client-secret=$(GITHUB_CLIENT_SECRET)

.PHONY: gen
gen: ## Runs go generate
	@go generate ./...

.PHONY: lint
lint: ## Runs staticcheck linter
	GOFLAGS=-buildvcs=false go tool staticcheck ./...

.PHONY: test
test: test-mock test-integration ## Runs all tests

.PHONY: test-mock
test-mock: ## Runs unit/mock tests (no services needed)
	go test ./... -timeout 120s -coverprofile=coverage.out

.PHONY: test-integration
test-integration: ## Runs integration tests with in-memory SQLite
	@PKREG_TEST_DB_SYSTEMS=$${PKREG_TEST_DB_SYSTEMS:-mem,sqlite} \
	go test -tags integration ./integration/

.PHONY: test-services-up
test-services-up: ## Start test services (PostgreSQL)
	@docker-compose -f docker/docker-compose.yml up -d postgresql

.PHONY: test-services-down
test-services-down: ## Stop test services
	@docker-compose -f docker/docker-compose.yml down -v --remove-orphans

.PHONY: test-fe
test-fe: ## Runs frontend integration tests with Selenium
	@go test -tags integration -run TestFrontend ./integration/ -timeout 300s

.PHONY: test-backends
test-backends: ## Runs integration tests with all backends (requires test-services-up)
	@PKREG_TEST_DB_SYSTEMS=mem,sqlite,postgresql \
	go test -tags integration ./integration/backends/... -coverprofile=coverage-backends.out

.PHONY: tag
tag: ## Tag a release: make tag SEMVER=major|minor|patch
ifndef SEMVER
	$(error SEMVER is required. Usage: make tag SEMVER=major|minor|patch)
endif
ifeq ($(filter $(SEMVER),major minor patch),)
	$(error SEMVER must be one of: major, minor, patch)
endif
	$(eval CURRENT := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0))
	$(eval MAJOR := $(shell echo $(CURRENT) | sed 's/^v//' | cut -d. -f1))
	$(eval MINOR := $(shell echo $(CURRENT) | sed 's/^v//' | cut -d. -f2))
	$(eval PATCH := $(shell echo $(CURRENT) | sed 's/^v//' | cut -d. -f3))
ifeq ($(SEMVER),major)
	$(eval NEXT := v$(shell echo $$(($(MAJOR)+1))).0.0)
else ifeq ($(SEMVER),minor)
	$(eval NEXT := v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0)
else ifeq ($(SEMVER),patch)
	$(eval NEXT := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))))
endif
	$(eval NEXT_BARE := $(shell echo $(NEXT) | sed 's/^v//'))
	sed -i 's/^## \[Unreleased\]$$/## [Unreleased]\n\n## [$(NEXT_BARE)] - $(shell date +%Y-%m-%d)/' CHANGELOG.md
	git add CHANGELOG.md
	git commit -m "Release $(NEXT)"
	git tag -a $(NEXT) -m "Release $(NEXT)"
	git push origin main $(NEXT)
	@echo ""
	@echo "===== Released $(NEXT) ====="

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -X github.com/pikoci/registry/cmd.Version=$(VERSION) -X github.com/pikoci/pkreg/cmd.Commit=$(COMMIT)

.PHONY: release $(PLATFORMS)
release: $(PLATFORMS) ## Creates the bin on the ./builds/

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags "$(LDFLAGS)" -o ./builds/'$(os)-$(arch)' .
