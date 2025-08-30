# Build worker
FROM golang:1.23 as gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

# Allure CLI layer
FROM node:20-slim as allure
RUN npm install -g allure-commandline@2.29.0

# Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates unzip zstd tar curl sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=allure /usr/local/lib/node_modules/allure-commandline /opt/allure-commandline
RUN ln -s /opt/allure-commandline/bin/allure /usr/local/bin/allure
COPY --from=gobuild /out/worker /usr/local/bin/worker
WORKDIR /work
VOLUME ["/reports","/data"]
ENTRYPOINT ["/usr/local/bin/worker"]
