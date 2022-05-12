FROM golang:1.18-alpine AS builder

RUN apk add --no-cache git

WORKDIR /workspace

COPY . .

RUN go build -o bin/main main.go

RUN sh ./scripts/prepare_env.sh

FROM alpine:3.13 as production

WORKDIR /bin

RUN apk add --no-cache bash

COPY --from=builder /workspace/bin/main main

COPY --from=builder /workspace/app/dao/sqldb/migrations/ app/dao/sqldb/migrations/

COPY --from=builder /workspace/.repo.env .repo.env

CMD ["/bin/main"]
