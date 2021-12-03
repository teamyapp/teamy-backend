FROM alpine:3.13 as production

WORKDIR /app

RUN apk add --no-cache bash

COPY bin/main bin/main

COPY .repo.env .repo.env

CMD ["/app/bin/main"]
