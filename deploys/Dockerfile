
FROM golang:1.22-alpine3.18 AS builder

# Update packages and install needed dependencies
RUN apk update && \
    apk add ca-certificates git curl gcc musl-dev build-base autoconf automake libtool

WORKDIR /app

COPY . .

ENV GOFLAGS=-buildvcs=false

RUN go build -o /app/main .

CMD ["/app/main"]
