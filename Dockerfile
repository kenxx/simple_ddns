FROM golang:1.18-alpine as build

WORKDIR /data

COPY . /data/

RUN go build ./cmd/simple_ddns.go

FROM alpine:3.16

COPY --from=build /data/simple_ddns /usr/local/bin/

WORKDIR /data

CMD simple_ddns