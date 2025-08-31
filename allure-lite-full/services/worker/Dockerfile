# Build worker
FROM golang:1.23 as gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

# Allure CLI layer
FROM openjdk:17-alpine AS allure
WORKDIR /opt
ARG ALLURE_VERSION
ENV ALLURE_VERSION=${ALLURE_VERSION}
RUN wget -qO- https://github.com/allure-framework/allure2/releases/download/${ALLURE_VERSION}/allure-${ALLURE_VERSION}.tgz | tar -xz

# Runtime
FROM debian:bookworm-slim
ARG ALLURE_VERSION
ENV ALLURE_VERSION=${ALLURE_VERSION}

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates unzip zstd tar curl sqlite3 && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y openjdk-17-jre && \
    rm -rf /var/lib/apt/lists/*

ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH="${JAVA_HOME}/bin:${PATH}"

# Copy allure from allure stage
COPY --from=allure /opt/allure-${ALLURE_VERSION} /opt/allure-${ALLURE_VERSION}
ENV PATH="/opt/allure-${ALLURE_VERSION}/bin:$PATH"

COPY --from=gobuild /out/worker /usr/local/bin/worker
WORKDIR /work
VOLUME ["/reports","/data"]
ENTRYPOINT ["/usr/local/bin/worker"]
