# syntax=docker/dockerfile:1

FROM golang:1.21
WORKDIR /allure-server

COPY go.mod go.sum ./

RUN go mod download

COPY app/*.go  ./app/
COPY config/*.go  ./config/
COPY utils/*.go  ./utils/
COPY globals/*.go  ./globals/
COPY token/*.go  ./token/
COPY swagger/*.go  ./swagger/
COPY static/*  ./static/


COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-allure-server-go
# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 8080

# Run
CMD ["/docker-allure-server-go"]