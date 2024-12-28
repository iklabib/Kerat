FROM alpine:3.20
WORKDIR /workspace
RUN apk add --no-cache icu-libs
COPY containerfiles/entry.sh .
ENTRYPOINT [ "sh", "/workspace/entry.sh" ]