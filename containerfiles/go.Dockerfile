FROM registry.access.redhat.com/ubi9/ubi:9.4 AS build
ENV VERSION="1.23.2"
ENV PATH="/usr/local/go/bin:$PATH"
ADD https://go.dev/dl/go${VERSION}.linux-amd64.tar.gz go.tar.gz
RUN tar -xf go.tar.gz && mv go /usr/local && rm go.tar.gz