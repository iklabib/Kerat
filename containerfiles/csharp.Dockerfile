FROM alpine:3.20 AS build
RUN apk add --no-cache dotnet8-sdk
ADD https://github.com/iklabib/LittleRosie/archive/refs/heads/master.zip csharp.zip
RUN unzip csharp.zip && mv LittleRosie-master build
WORKDIR /build
RUN dotnet publish -r linux-x64 -o output LittleRosie.csproj

FROM kerat:alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["/usr/bin/evaluator", "/app/LittleRosie"]

