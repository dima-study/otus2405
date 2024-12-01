PATH := $(shell go env GOPATH)/bin:$(PATH)

VERSION := develop
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

CALENDAR_BIN := "bin/calendar"
CALENDAR_CONFIG := "configs/calendar.yaml"

SCHEDULER_BIN := "bin/scheduler"
SCHEDULER_CONFIG := "configs/scheduler.yaml"

SENDER_BIN := "bin/sender"
SENDER_CONFIG := "configs/sender.yaml"

DOCKER_IMG_PG = "calendar-pg:$(VERSION)"
DOCKER_IMG_AMQP = "calendar-amqp:$(VERSION)"
DOCKER_IMG_CALENDAR := "calendar-calendar:$(VERSION)"
DOCKER_IMG_SCHEDULER := "calendar-scheduler:$(VERSION)"
DOCKER_IMG_SENDER := "calendar-sender:$(VERSION)"

PG_HOST = 127.0.0.1
PG_USER = calendar
PG_DB = calendar
PG_PSWD = calendar
