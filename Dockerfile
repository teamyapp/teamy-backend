FROM golang:1.17-alpine AS builder

WORKDIR /app

RUN apk add git bash

# Install dependencies
COPY go.mod go.sum ./

ARG GITHUB_USERNAME

ARG GITHUB_PERSONAL_ACCESS_TOKEN

RUN git config --global url."https://${GITHUB_USERNAME}:${GITHUB_PERSONAL_ACCESS_TOKEN}@github.com".insteadOf "https://github.com"

RUN go env -w GOPRIVATE=github.com/teamyapp/*

RUN go mod download

# Verify dependencies
RUN go mod verify

COPY . .

RUN go build -o bin/main main.go

FROM alpine:3.13 as production

WORKDIR /app

RUN apk add --no-cache bash

COPY --from=builder /app/bin/main ./bin/main

CMD ["/app/bin/main"]
