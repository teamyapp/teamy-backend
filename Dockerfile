FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o bin/main main.go

FROM alpine:3.13 as production

WORKDIR /app

RUN apk add --no-cache bash

COPY --from=builder /app/bin/main ./bin/main

CMD ["/app/bin/main"]
