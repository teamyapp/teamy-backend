FROM alpine:3.13 as production

WORKDIR /workspace

RUN apk add --no-cache bash

COPY app/repo/db/migration/ app/repo/db/migration/

COPY bin/main main

COPY .repo.env .repo.env

CMD ["/workspace/main"]
