FROM alpine:3.20
RUN apk add --no-cache icu-libs
WORKDIR /workspace
ENTRYPOINT [ "sh", "chmod", "+x", "/workspace/box", "&&", "/workspace/box"]