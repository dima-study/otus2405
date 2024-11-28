# Build
FROM golang:1.22-bookworm AS build

RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      protobuf-compiler \
      libprotobuf-dev

RUN mkdir -p /opt/calendar
WORKDIR /opt/calendar

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN make build-app-scheduler

# Release
FROM alpine:3

COPY --from=build /opt/calendar/bin/scheduler /bin/calendar-scheduler
COPY --from=build /opt/calendar/configs/scheduler.yaml /etc/calendar/scheduler.yaml

CMD ["/bin/calendar-scheduler", "-config", "/etc/calendar/scheduler.yaml"]
