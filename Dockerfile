FROM golang:1.23-bullseye

RUN apt-get update  &&  go install github.com/cosmtrek/air@v1.52.1

# Set working directory
WORKDIR /app

EXPOSE 8080