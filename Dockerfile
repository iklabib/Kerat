FROM golang:1.24-bookworm AS build
WORKDIR /build
ARG TARGETARCH
ENV GOOS="linux"
ENV GOARCH=$TARGETARCH 

RUN apt update && apt install -y ca-certificates
COPY ["go.sum", "go.mod", "./"]
RUN go mod download

COPY . .
RUN go build -o kerat cmd/kerat/main.go

FROM debian:bookworm-slim AS final
# RUN apt update && apt install -y musl-dev libicu74 libicu-dev git curl aria2
RUN apt update && apt install -y libicu72 git curl aria2
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
