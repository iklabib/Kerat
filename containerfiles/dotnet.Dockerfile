FROM debian:bookworm-slim AS package

# install ICU package for .NET globalization
# we can skip this by disabling globalization on runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    libicu72 \
    && rm -rf /var/lib/apt/lists/*

FROM gcr.io/distroless/base-debian12:nonroot AS final

COPY --from=package /usr/lib/x86_64-linux-gnu/libicudata.so.72* /usr/lib/x86_64-linux-gnu/
COPY --from=package /usr/lib/x86_64-linux-gnu/libicui18n.so.72* /usr/lib/x86_64-linux-gnu/
COPY --from=package /usr/lib/x86_64-linux-gnu/libicuio.so.72* /usr/lib/x86_64-linux-gnu/
COPY --from=package /usr/lib/x86_64-linux-gnu/libicutest.so.72* /usr/lib/x86_64-linux-gnu/
COPY --from=package /usr/lib/x86_64-linux-gnu/libicutu.so.72* /usr/lib/x86_64-linux-gnu/
COPY --from=package /usr/lib/x86_64-linux-gnu/libicuuc.so.72* /usr/lib/x86_64-linux-gnu/

WORKDIR /workspace
