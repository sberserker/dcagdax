FROM golang:1.16.2-alpine3.13 as build

WORKDIR /src/

COPY go.mod . 
COPY go.sum .
RUN go mod download

COPY *.go .
COPY coinbase coinbase

RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/  ./...

FROM alpine:3.13.2

RUN apk update && \
    apk add --no-cache tzdata

COPY --from=build /go/bin/dcagdax /bin/dcagdax

# copy crontabs for root user
COPY cron.conf /etc/crontabs/root

# start crond with log level 8 in foreground, output to stderr
CMD ["crond", "-f", "-d", "8"]