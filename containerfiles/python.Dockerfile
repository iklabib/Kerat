FROM kerat:alpine
WORKDIR /box
USER root
RUN apk add --no-cache python3
USER user
ENTRYPOINT ["/usr/bin/evaluator", "python"]