FROM registry.access.redhat.com/ubi9/ubi:9.4 AS build
ADD https://go.dev/dl/go1.23.2.linux-amd64.tar.gz go.tar.gz
RUN tar -xf go.tar.gz && mv go /usr/local && rm go.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

WORKDIR /build
COPY . .
RUN go build -ldflags "-s -w" -o evaluator cmd/eval/main.go

FROM registry.access.redhat.com/ubi9/ubi-micro:9.4
COPY --from=build /build/evaluator /usr/bin/evaluator
ENTRYPOINT [ "/usr/bin/evaluator" ]