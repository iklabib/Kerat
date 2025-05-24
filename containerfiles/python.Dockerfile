FROM gcr.io/distroless/python3-debian12:nonroot

WORKDIR /kerat
COPY template/python .
WORKDIR /workspace
ENTRYPOINT ["python3", "/kerat/main.py" ]
