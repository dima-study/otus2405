FROM alpine:3

RUN apk update && \
    apk add which && \
    apk add make && \
    rm -rf /var/cache/apk/*

ADD https://github.com/pressly/goose/releases/download/v3.22.0/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

WORKDIR /app

COPY Makefile .
COPY migrations ./migrations

CMD ["make", "migrate-pg"]
