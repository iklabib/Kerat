FROM golang:1.23.3-alpine3.20 AS build
WORKDIR /build
COPY . .
RUN go build -ldflags "-s -w" -o evaluator cmd/eval/main.go

FROM alpine:3.20
RUN apk add --no-cache icu-libs
COPY --from=build /build/evaluator /usr/bin/evaluator
ENTRYPOINT [ "/usr/bin/evaluator" ]