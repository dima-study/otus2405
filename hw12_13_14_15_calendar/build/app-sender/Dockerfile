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
RUN make build-app-sender

# Release
FROM alpine:3

COPY --from=build /opt/calendar/bin/sender /bin/calendar-sender
COPY --from=build /opt/calendar/configs/sender.yaml /etc/calendar/sender.yaml

CMD ["/bin/calendar-sender", "-config", "/etc/calendar/sender.yaml"]
