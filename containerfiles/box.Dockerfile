FROM kerat:go AS build
WORKDIR /build
COPY . .
RUN go build -ldflags "-s -w" -o evaluator cmd/eval/main.go

FROM registry.access.redhat.com/ubi9/ubi-micro:9.4
COPY --from=build /build/evaluator /usr/bin/evaluator
ENTRYPOINT [ "/usr/bin/evaluator" ]