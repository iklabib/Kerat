FROM ubuntu:noble AS build
WORKDIR /build
ARG TARGETARCH
ENV GO_VERSION="1.23.3"
ENV PATH="/usr/local/go/bin:$PATH"
ADD https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz go.tar.gz
RUN tar -xf go.tar.gz && rm -rf /usr/local/go && mv go /usr/local && rm go.tar.gz

RUN apt update && apt install -y ca-certificates
COPY . .
RUN go build -o kerat cmd/kerat/main.go

FROM ubuntu:noble AS final
RUN apt update && apt install -y musl-dev libicu74 libicu-dev git curl aria2
ARG TARGETARCH

# copy templates
ENV REPOSITORY=/repository
COPY template ${REPOSITORY}

# csharp
ENV DOTNET_VERSION="8.0.404"
RUN ${REPOSITORY}/csharp/setup.sh ${DOTNET_VERSION} ${TARGETARCH}

RUN apt clean && rm -rf /var/lib/apt/lists/*

ENV PATH="/root/dotnet:$PATH"
WORKDIR /app
COPY --from=build /build/kerat .
COPY config.yaml config.yaml
ENTRYPOINT ["/app/kerat"]