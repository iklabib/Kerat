FROM python:3.13-alpine3.20

WORKDIR /app
COPY template/python .
ENTRYPOINT ["python3", "/app/main.py" ]