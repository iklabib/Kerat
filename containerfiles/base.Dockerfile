FROM golang:bookworm AS build
WORKDIR /build
COPY . .
RUN go build -ldflags "-s -w" -o evaluator cmd/eval/main.go

FROM ubuntu:noble
COPY --from=build /build/evaluator /usr/bin/evaluator
