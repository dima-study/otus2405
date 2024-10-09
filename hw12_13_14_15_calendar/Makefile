ifneq (,$(wildcard ./.env.mk))
    include .env.mk
    export
endif

.PHONY: help
help:
	@printf "%-20s %s\n" "Target" "Description"
	@printf "%-20s %s\n" "------" "-----------"
	@make -pqR : 2>/dev/null \
		| awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' \
    | sort \
    | grep -v -e '^[^[:alnum:]]' -e '^$@$$' \
    | xargs -I _ sh -c 'printf "%-20s " _; make _ -nB | (grep -i "^# Help:" || echo "") | tail -1 | sed "s/^# Help: //g"'


#
# GENERATE
#
.PHONY: generate
generate: generate-proto
	@# Help: Do generate sources


.PHONY: generate-proto
generate-proto: install-deps-protoc
	@# Help: Generate from proto files
	protoc \
		-Iapi/vendor/proto/alta/protopatch/ \
		-Iapi/vendor/proto/googleapis/googleapis \
		-Iapi/proto \
		--go-patch_out=plugin=go,paths=source_relative:internal/api/proto \
		--go-patch_out=plugin=go-grpc,paths=source_relative,require_unimplemented_servers=true:internal/api/proto \
		--go-patch_out=plugin=grpc-gateway,paths=source_relative:internal/api/proto \
		--openapiv2_out=output_format=yaml,allow_merge=true,merge_file_name=api,json_names_for_fields=false,openapi_naming_strategy=simple:api/openapi-spec \
		./api/proto/event/v1/*.proto


#
# BUILD
#

.PHONY: build
build:
	@# Help: build calendar app $(CALENDAR_BIN)
	go build -v -o $(CALENDAR_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar


.PHONY: build-img-all
build-img-all: build-img-pg
	@# Help: Build all docker images


.PHONY: build-img-pg
build-img-pg:
	@# Help: Build postgres docker image $(DOCKER_IMG_PG)
	docker build \
		-t $(DOCKER_IMG_PG) \
		-f build/pg/Dockerfile build/pg


#
# RUN
#

.PHONY: run
run: build
	@# Help: Run calendar app $(CALENDAR_BIN) with default config $(CALENDAR_CONFIG)
	@$(CALENDAR_BIN) -config $(CALENDAR_CONFIG)


.PHONY: run-img-pg
run-img-pg: build-img-pg
	@# Help: Run temp posgtres docker container from $(DOCKER_IMG_PG) image
	docker run --rm \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		-e PG_USER=$(PG_USER) \
		-e PG_DB=$(PG_DB) \
		-e PG_PSWD=$(PG_PSWD) \
		-p 5432:5432 \
		$(DOCKER_IMG_PG)


#
# Deps
#

.PHONY: install-deps-all
install-deps-all: install-deps-migrate install-deps-lint
	@# Help: Install all deps

.PHONY: install-deps-migrate
install-deps-migrate:
	@# Help: Install goose migration tool
	@(which goose 1>/dev/null 2>&1) && echo "goose migration tool is already installed" \
		|| ( echo "install goose migration tool..." \
				&& go install github.com/pressly/goose/v3/cmd/goose@latest \
				&& echo "done" \
			)

.PHONY: install-deps-lint
install-deps-lint:
	@# Help: Install golangci linter
	@(which golangci-lint 1>/dev/null 2>&1) && echo "golangci-lint is already installed" \
		|| ( echo "install golangci-lint..." \
				&& curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2 \
				&& echo "done" \
			)


#
# Util commands
#

.PHONY: migrate-pg
migrate-pg: install-deps-migrate
		@# Help: Migrate UP postgres db
		goose \
			-dir ./migrations/pg \
			postgres "postgresql://$(PG_USER):$(PG_PSWD)@$(PG_HOST):5432/$(PG_DB)?sslmode=disable" \
			up


.PHONY: test
test:
	@# Help: Run tests
	go test -race ./internal/... ./cmd/...


.PHONY: lint
lint: install-deps-lint
	@# Help: Lint the project
	golangci-lint run --config=.golangci.yml ./...
