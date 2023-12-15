FROM golang:1.21.5-alpine3.19 as build

WORKDIR /src/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go .
COPY clients clients
COPY exchanges exchanges

RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/  ./...

FROM alpine:3.19.0

RUN apk update && \
    apk add --no-cache tzdata

COPY --from=build /go/bin/dcagdax /bin/dcagdax

# copy crontabs for root user
COPY cron.conf /etc/crontabs/root

# start crond with log level 8 in foreground, output to stderr
CMD ["crond", "-f", "-d", "8"]