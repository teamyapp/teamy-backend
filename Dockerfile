FROM golang:1.18-alpine AS builder

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

RUN go mod verify

COPY . .

RUN go build -o bin/main main.go

RUN sh ./scripts/prepare_env.sh

FROM alpine:3.13 as production

WORKDIR /bin

RUN apk add --no-cache bash

COPY --from=builder /workspace/bin/main main

COPY --from=builder /workspace/migrations/ migrations/

COPY --from=builder /workspace/.repo.env .repo.env

CMD ["/bin/main"]
