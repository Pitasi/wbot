# Build the application from source
FROM golang:1.20-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /wbot

FROM alpine:3.17

WORKDIR /

COPY --from=build-stage /wbot /wbot

ENTRYPOINT ["/wbot"]
