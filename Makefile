DB_URL="root:1550@/backend-portal?parseTime=true"

WORKER_IMAGE = golang:1.21-alpine3.19
OAS3_GENERATOR_DOCKER_IMAGE = openapitools/openapi-generator-cli:latest-release

.PHONY: setup
setup:
	git config --local core.hooksPath githooks/

.PHONY: ci-swag
ci-swag:
	@docker run --rm -v $(PWD):/work $(WORKER_IMAGE) \
		sh -c "apk update && apk add --no-cache git; go install github.com/swaggo/swag/cmd/swag@latest && (cd /work; swag init)"

.PHONY: ci-swaggen
ci-swaggen: ci-swag
	@echo "[OAS3] Converting Swagger 2-to-3 (yaml)"
	@docker run --rm -v $(PWD)/docs:/work $(OAS3_GENERATOR_DOCKER_IMAGE) \
	  generate -i /work/swagger.yaml -o /work/v3 -g openapi-yaml --minimal-update
	@docker run --rm -v $(PWD)/docs/v3:/work $(WORKER_IMAGE) \
	  sh -c "rm -rf /work/.openapi-generator"
	@echo "[OAS3] Copying openapi-generator-ignore (json)"
	@docker run --rm -v $(PWD)/docs/v3:/work $(WORKER_IMAGE) \
	  sh -c "cp -f /work/.openapi-generator-ignore /work/openapi"
	@echo "[OAS3] Converting Swagger 2-to-3 (json)"
	@docker run --rm -v $(PWD)/docs:/work $(OAS3_GENERATOR_DOCKER_IMAGE) \
	  generate -s -i /work/swagger.json -o /work/v3/openapi -g openapi --minimal-update
	@echo "[OAS3] Cleaning up generated files"
	@docker run --rm -v $(PWD)/docs:/work $(WORKER_IMAGE) \
	   sh -c "mv -f /work/v3/openapi/openapi.json /work/swagger.json ; mv -f /work/v3/openapi/openapi.yaml /work/swagger.yaml ; rm -rf /work/v3"

.PHONY: swagger
swagger:
	@echo "Generating swagger documentation"
		swag init --ot yaml,json
	@echo "[OAS3] Converting Swagger 2-to-3 (yaml)"
		openapi-generator generate -i ./docs/swagger.yaml -o ./docs/v3 -g openapi-yaml --minimal-update
		rm -rf ./docs/v3/.openapi-generator
	@echo "[OAS3] Copying openapi-generator-ignore (json)"
		cp -f ./docs/v3/.openapi-generator-ignore ./docs/v3/openapi/
	@echo "[OAS3] Converting openapi.json to openapi.yaml"
		openapi-generator generate -s -i ./docs/swagger.json -o ./docs/v3/openapi -g openapi --minimal-update
	@echo "[OAS3] Cleaning up generated files"
		mv -f ./docs/v3/openapi/openapi.json ./docs/swagger.json
		mv ./docs/v3/openapi/openapi.yaml ./docs/swagger.yaml
		rm -rf ./docs/v3/

# Migration
MIGRATION_NAME ?= filename

.PHONY: migrate-up
migrate-up:
	@echo ">>> executing UP migration..."
	@go run main.go migrate up
	@echo ">>> finished executing migration..."

.PHONY: migrate-down	
migrate-down:
	@echo ">>> executing DOWN migration..."
	@go run main.go migrate down
	@echo ">>> finished executing migration..."

.PHONY: migrate-status
migrate-status:
	@echo ">>> executing STATUS migration..."
	@go run main.go migrate status

.PHONY: new-migration
new-migration:
	@echo ">>> Create SQL Migration"
	@go run main.go migrate create $(MIGRATION_NAME) sql

.PHONY: gen-mock-repo
gen-mock-repo:
	mockery --all --dir ./internal/repository/ --output ./mocks/repository --outpkg mocks
	mockery --all --dir ./internal/repository/datamart --output ./mocks/repository --outpkg mocks

.PHONY: gen-mock-service
gen-mock-service:
	mockery --all --dir ./internal/service/ --output ./mocks/service --outpkg mocks

.PHONY: gen-mock-controller
gen-mock-controller:
	mockery --all --dir ./port/http/controller/ --output ./mocks/port/http/controller --outpkg mocks

.PHONY: gen-pkg-mock
gen-pkg-mock:
	mockery --all --dir ./pkg/logger/ --output ./mocks/pkg/logger --outpkg mocks
	mockery --all --dir ./pkg/mySqlExt/ --output ./mocks/pkg/mySqlExt --outpkg mocks
	mockery --all --dir ./pkg/redisExt/ --output ./mocks/pkg/redisExt --outpkg mocks
	mockery --all --dir ./pkg/snapCore/ --output ./mocks/pkg/snapCore --outpkg mocks
	mockery --all --dir ./pkg/rabbitMqExt/ --output ./mocks/pkg/rabbitmqExt --outpkg mocks
	mockery --all --dir ./pkg/httpRequestExt/ --output ./mocks/pkg/httpRequestExt --outpkg mocks
	mockery --all --dir ./pkg/callback/ --output ./mocks/pkg/callback --outpkg mocks
	mockery --all --dir ./pkg/gcs/ --output ./mocks/pkg/gcs --outpkg mocks
	mockery --all --dir ./pkg/encryption/ --output ./mocks/pkg/encryption --outpkg mocks
	mockery --all --dir ./pkg/slackExt/ --output ./mocks/pkg/slackExt --outpkg mocks
	mockery --all --dir ./pkg/jwt/ --output ./mocks/pkg/jwt --outpkg mocks
	mockery --all --dir ./pkg/xlsx/ --output ./mocks/pkg/xlsx --outpkg mocks
	mockery --all --dir ./pkg/tablePartitionExt/ --output ./mocks/pkg/tablePartitionExt --outpkg mocks
	mockery --all --dir ./pkg/fds/ --output ./mocks/pkg/fds --outpkg mocks
	mockery --name PDFGenerator --dir ./pkg/pdf/ --output ./mocks/pkg/pdf --outpkg mocks
	mockery --name ILogger --dir ./pkg/ --output ./mocks/pdk/logger --outpkg mocks
	mockery --name IGCPSecretManager --dir ./pkg/ --output ./mocks/pdk/gcp --outpkg mocks
	mockery --name Encrypter --dir ./pkg/ --output ./mocks/pdk/encrypt --outpkg mocks
	mockery --all --dir ./pkg/vault/ --output ./mocks/pkg/vault --outpkg mocks

.PHONY: gen-mocks
gen-mocks: gen-mock-repo gen-mock-service gen-mock-controller gen-pkg-mock

# Run the application
.PHONY: run-web
run-web:
	go run main.go serveWeb --config .config.yaml --secret .secret.yaml
.PHONY: run
run-http:
	go run main.go serveHttp --config .config.yaml --secret .secret.yaml

# Run the consumer
.PHONY: run-consumer
run-consumer:
	go run main.go serveConsumer --config .config.yaml --secret .secret.yaml

# Run the cronJob manually
.PHONY: run-cron
run-cron:
	go run main.go serveCron --config .config.yaml --secret .secret.yaml --cronJob=$(command) --date="$(date)"

# Run the console command
.PHONY: run-console
run-console:
	go run main.go serveConsole --config .config.yaml --secret .secret.yaml --command=$(command) --startDate="$(startDate)" --endDate="$(endDate)"

# build with docker
.PHONY: build
build:
	docker build -t backend-portal .

# run unit tests
.PHONY: run-test
run-test:
	@echo "Running unit tests"
	@./scripts/run-unit-tests.sh

# run integration tests
.PHONY: run-integration-test
run-integration-test:
	@echo "Running integration tests"
	@INTEGRATION_TEST=1 go tool gotestsum --format=testdox --format-hide-empty-pkg -- -run=TestIntegration ./...

.PHONY: run-all-test
run-all-test:
	@echo "Running unit tests and integration tests"
	@INTEGRATION_TEST=1 go test -race ./...

# can be added with generate open api command
.PHONY: generate-open-api
generate-open-api:
	openapi-generator generate -i docs/swagger.yaml -g html2 -o docs/swagger
	openapi-generator generate -i docs/merchant-rcn.yaml -g html2 -o docs/merchant-rcn

url := input
TEMP_DOWNLOAD_DIR := ./temp-download-$(url)
proto-compile:
	mkdir -p $(TEMP_DOWNLOAD_DIR)
	fetcher --url=$(url) --out=$(TEMP_DOWNLOAD_DIR) --file=./downloader_config.json
	protoc $(TEMP_DOWNLOAD_DIR)/*.proto --go_out=./internal/model/proto/
	rm -rf $(TEMP_DOWNLOAD_DIR)

## Compile Proto Files for Golang ##
# Input
proto_direct ?=
proto_temp ?= /tmp
proto_branch ?= main
proto_service ?= backend-portal

# Variable
proto_output := ./tmp
proto_domain := github.com/paper-indonesia
proto_folder := $(proto_temp)/$(proto_domain)/proto
proto_git := git@github.com:paper-indonesia/proto.git
proto_common := $(proto_domain)/backend-portal/internal/model/proto/common
proto_dest := ./internal/model/proto

# Main commands
.PHONY: run-protoc
run-protoc:
	@mkdir -p $(proto_output)
	@if [ "$(proto_direct)" != "" ]; then\
		make proto-compile-offline proto_direct=$(proto_direct) proto_service=$(proto_service);\
	else\
		make proto-compile-online proto_temp=$(proto_temp) proto_branch=$(proto_branch) proto_service=$(proto_service);\
	fi
	@mkdir -p $(proto_dest)
	@cp -rf ./tmp/$(proto_common) $(proto_dest)/
	@cp -rf  ./tmp/$(proto_domain)/proto/payment-gateway/$(proto_service)/* $(proto_dest)/
	@rm -R -fd $(proto_output)

# Subprocess for offline compilation process
.PHONY: proto-compile-offline
proto-compile-offline:
	@echo ""
	@echo "Compile from local repository..."
	@protoc --proto_path=$(proto_direct) \
	 	--go_out=$(proto_output) \
		--go_opt=Mpayment-gateway/$(proto_service)/common/amount.proto=$(proto_common) \
	 	$(proto_direct)/payment-gateway/$(proto_service)/*/*/*.proto \
		$(proto_direct)/payment-gateway/$(proto_service)/*/*.proto
	@echo ""

# Subprocess for online compilation process
.PHONY: proto-compile-online
proto-compile-online:
	@echo ""
	@echo "Proto repository pull..."
	@if [ ! -d $(proto_folder) ]; then\
		mkdir -p $(proto_folder);\
		git clone $(proto_git) $(proto_folder);\
	fi
	@cd $(proto_folder) && git pull origin $(proto_branch) && git checkout $(proto_branch)
	@protoc --proto_path=$(proto_folder) \
	 	--go_out=$(proto_output) \
		--go_opt=Mpayment-gateway/$(proto_service)/common/amount.proto=$(proto_common) \
	 	$(proto_folder)/payment-gateway/$(proto_service)/*/*/*.proto \
		$(proto_folder)/payment-gateway/$(proto_service)/*/*.proto
	@echo ""

.PHONY: lint lint-summary

# Run the linter and save the report
lint:
	@cd ./scanner-config/ \
		&& golangci-lint custom -v \
		&& mv custom-gcl ../ \
		&& cd ../
	@echo "Running custom golangci-lint"
	@./custom-gcl run --config ./scanner-config/.golangci.yaml --out-format json > lint-report.json || echo "Lint issues detected."
	@echo "\n=== Lint Errors ==="
	@./custom-gcl run --config ./scanner-config/.golangci.yaml || true
	@rm ./custom-gcl

# Generate the lint summary
lint-summary:
	@echo "\n=== Lint Summary ==="
	@total_issues=$$(grep -o '"Pos":' lint-report.json | wc -l | tr -d '[:space:]'); \
	total_files=$$(find . -name "*.go" | wc -l | tr -d '[:space:]'); \
	if [ "$$total_files" -gt 0 ]; then \
	  percentage=$$((100 - (100 * $$total_issues / $$total_files))); \
	  echo "Lint Passed: $$percentage% (Issues: $$total_issues / Files: $$total_files)"; \
	  if [ "$$percentage" -lt 80 ]; then \
	    echo "Failing build: Lint pass percentage below threshold (80%)."; \
	    exit 1; \
	  fi; \
	else \
	  echo "No Go files found."; \
	  exit 0; \
	fi
