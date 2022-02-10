FROM golang:1.17.6-bullseye

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ONBUILD ARG BENCH
ONBUILD RUN go build -tags ${BENCH} -v -o /usr/local/bin/gobench ./...

