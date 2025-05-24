FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /workspace
ENTRYPOINT [ "/workspace/box" ]
