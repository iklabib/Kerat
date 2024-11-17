FROM ubuntu:noble AS build
WORKDIR /build
ENV VERSION="1.23.3"
ENV PATH="/usr/local/go/bin:$PATH"
ADD https://go.dev/dl/go${VERSION}.linux-amd64.tar.gz go.tar.gz
RUN tar -xf go.tar.gz && mv go /usr/local && rm go.tar.gz
COPY . .
RUN go build -ldflags "-s -w" -o evaluator cmd/eval/main.go

FROM ubuntu:noble
COPY --from=build /build/evaluator /usr/bin/evaluator
ENTRYPOINT [ "/usr/bin/evaluator" ]