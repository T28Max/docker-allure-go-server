# Build
FROM golang:1.23 as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

# Run
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/api /app/api
ENV TZ=UTC
EXPOSE 8081
VOLUME ["/data","/reports"]
ENTRYPOINT ["/app/api"]
