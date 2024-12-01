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

CMD ["go", "test", "./tests/...", "-v", "-count=1", "-tags=integration"]
