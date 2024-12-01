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
		-Iapi/vendor/proto/grpc-ecosystem/grpc-gateway \
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
build: build-app-calendar build-app-scheduler build-app-sender
	@# Help: build all apps


.PHONY: build-app-calendar
build-app-calendar: generate
	@# Help: build calendar app $(CALENDAR_BIN)
	CGO_ENABLED=0 go build -v -o $(CALENDAR_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar


.PHONY: build-app-scheduler
build-app-scheduler:
	@# Help: build scheduler app $(SCHEDULER_BIN)
	CGO_ENABLED=0 go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/scheduler


.PHONY: build-app-sender
build-app-sender:
	@# Help: build sender app $(SENDER_BIN)
	CGO_ENABLED=0 go build -v -o $(SENDER_BIN) -ldflags "$(LDFLAGS)" ./cmd/sender


.PHONY: build-img-all
build-img-all: build-img-pg build-img-amqp build-img-calendar build-img-scheduler build-img-sender
	@# Help: Build all docker images


.PHONY: build-img-calendar
build-img-calendar:
	@# Help: Build calender app docker image $(DOCKER_IMG_CALENDAR)
	docker build \
		-t $(DOCKER_IMG_CALENDAR) \
		-f build/app-calendar/Dockerfile .


.PHONY: build-img-scheduler
build-img-scheduler:
	@# Help: Build scheduler app docker image $(DOCKER_IMG_SCHEDULER)
	docker build \
		-t $(DOCKER_IMG_SCHEDULER) \
		-f build/app-scheduler/Dockerfile .


.PHONY: build-img-sender
build-img-sender:
	@# Help: Build sender app docker image $(DOCKER_IMG_SENDER)
	docker build \
		-t $(DOCKER_IMG_SENDER) \
		-f build/app-sender/Dockerfile .


.PHONY: build-img-pg
build-img-pg:
	@# Help: Build postgres docker image $(DOCKER_IMG_PG)
	docker build \
		-t $(DOCKER_IMG_PG) \
		-f build/pg/Dockerfile build/pg


.PHONY: build-img-amqp
build-img-amqp:
	@# Help: Build rabbitmq docker image $(DOCKER_IMG_AMQP)
	docker build \
		-t $(DOCKER_IMG_AMQP) \
		-f build/rabbitmq/Dockerfile build/rabbitmq

#
# RUN
#

.PHONY: up
up:
	@# Help: Run up-n-run docker with calendar project
	PG_HOST=pg docker-compose -f build/docker-compose.yaml up

.PHONY: down
down:
	@# Help: Stop docker with calendar project
	docker-compose -f build/docker-compose.yaml down

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


.PHONY: run-img-amqp
run-img-amqp: build-img-amqp
	@# Help: Run temp rabbitmq docker container from $(DOCKER_IMG_AMQP) image
	docker run --rm \
		-p 15672:15672 \
		-p 5672:5672 \
		$(DOCKER_IMG_AMQP)


#
# Deps
#

.PHONY: install-deps-all
install-deps-all: install-deps-migrate install-deps-lint install-deps-protoc
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
				&& curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.60.1 \
				&& echo "done" \
			)


.PHONY: install-deps-protoc
install-deps-protoc:
	@# Help: Instal protoc generator deps
	@ (which protoc-gen-go 1>/dev/null 2>&1) && echo "protoc-gen-go is already installed" \
		|| ( echo "install protoc-gen-go..." \
				&& go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35 \
				&& echo "done" \
			)
	@ (which protoc-gen-go-grpc 1>/dev/null 2>&1) && echo "protoc-gen-go-grpc is already installed" \
		|| ( echo "install protoc-gen-go-grpc..." \
				&& go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5 \
				&& echo "done" \
			)
	@ (which protoc-gen-grpc-gateway 1>/dev/null 2>&1) && echo "protoc-gen-grpc-gateway is already installed" \
		|| ( echo "install protoc-gen-grpc-gateway..." \
				&& go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.22 \
				&& echo "done" \
			)
	@ (which protoc-gen-openapiv2 1>/dev/null 2>&1) && echo "protoc-gen-openapiv2 is already installed" \
		|| ( echo "install protoc-gen-openapiv2..." \
				&& go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.22 \
				&& echo "done" \
			)
	@ (which protoc-gen-go-patch 1>/dev/null 2>&1) && echo "protoc-gen-go-patch is already installed" \
		|| ( echo "install protoc-gen-go-patch..." \
				&& go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@v0.5 \
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

.PHONY: integration-tests
integration-tests:
	@# Help: Run tests
	PG_HOST=pg ./tests/run.sh

.PHONY: lint
lint: install-deps-lint
	@# Help: Lint the project
	golangci-lint run --config=.golangci.yml ./...
