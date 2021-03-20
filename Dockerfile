FROM golang:1.16-alpine AS build-env

ENV CGO_ENABLED 0

RUN apk add --no-cache make git ncurses

WORKDIR /home/runner/work/gopass/gopass

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG golags_arg=""
ENV GOFLAGS=$goflags_arg

RUN make clean
RUN make gopass

FROM alpine:3.11
RUN apk add --no-cache ca-certificates git gnupg
COPY --from=build-env /home/runner/work/gopass/gopass/gopass /usr/local/bin/

