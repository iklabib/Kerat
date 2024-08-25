FROM golang:1.23.0-alpine3.20 AS build
WORKDIR /build
COPY . .
RUN go build -o evaluator cmd/eval/main.go

FROM alpine:3.20
RUN adduser -u 1000 -D user user
USER user
COPY --from=build /build/evaluator /usr/bin/evaluator
