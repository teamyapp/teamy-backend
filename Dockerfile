FROM alpine:3.13

WORKDIR /bin

RUN apk add --no-cache bash

COPY bin/main main

COPY doc/ doc/

COPY migrations/ migrations/

COPY .repo.env .repo.env

CMD ["/bin/main"]
