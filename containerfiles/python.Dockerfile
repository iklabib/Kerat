FROM python:3.13-alpine3.20

WORKDIR /kerat
COPY template/python .
RUN mkdir -p /workspace
ENTRYPOINT ["python3", "/kerat/main.py" ]