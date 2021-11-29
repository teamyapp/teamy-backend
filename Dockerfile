FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o bin/main main.go

RUN sh ./scripts/prepare_env.sh

FROM alpine:3.13 as production

WORKDIR /app

RUN apk add --no-cache bash

COPY --from=builder /app/bin/main ./bin/main

COPY --from=builder /app/.repo.env .repo.env

CMD ["/app/bin/main"]
