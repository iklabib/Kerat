FROM registry.access.redhat.com/ubi9/ubi:9.4

ENV VERSION=21.0.2
ADD https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-${VERSION}/graalvm-community-jdk-${VERSION}_linux-x64_bin.tar.gz /tmp/graalvm.tar.gz
RUN tar -xf /tmp/graalvm.tar.gz -C /tmp && \
    mkdir -p /usr/local && \
    mv /tmp/graalvm-community-openjdk-* /usr/local/graalvm && rm /tmp/graalvm.tar.gz

ENV PATH="/usr/local/graalvm/bin:$PATH"
ENTRYPOINT [ "/bin/sh" ]
