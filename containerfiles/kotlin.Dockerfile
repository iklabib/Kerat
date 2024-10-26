FROM kerat:java
ENV VERSION=2.0.21
ADD https://github.com/JetBrains/kotlin/releases/download/v${VERSION}/kotlin-native-prebuilt-linux-x86_64-${VERSION}.tar.gz /tmp/kotlin.tar.gz
RUN  tar -xf /tmp/kotlin.tar.gz -C /tmp && \
    mv /tmp/kotlin-native-prebuilt-linux-x86_64-* /usr/local/kotlin && \ 
    rm /tmp/kotlin.tar.gz

ENV PATH="/usr/local/kotlin/bin:$PATH"
# a hack to trigger dependecies download
RUN echo "fun main() {}" > /tmp/main.kt && kotlinc-native /tmp/main.kt -o /tmp/Main.class
ENTRYPOINT [ "bash" ]