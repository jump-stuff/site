# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.5

####
FROM golang:${GO_VERSION}-bookworm AS builder

RUN apt-get update -y && apt-get install -y build-essential
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY --exclude=.env.production --exclude=Dockerfile --exclude=.dockerignore --exclude=litefs.yml . .
RUN go build -v -o /run-app .

####
FROM debian:bookworm
WORKDIR /app

RUN apt-get update -y && apt-get install -y openssl ca-certificates fuse3 sqlite3 curl

COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs
COPY litefs.yml /etc/litefs.yml

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY db/migrations ./migrations/

COPY .env.production .
COPY --from=builder run-app /usr/local/bin/run-app

ENV JUMP_ENV=production
EXPOSE 8000
ENTRYPOINT litefs mount
