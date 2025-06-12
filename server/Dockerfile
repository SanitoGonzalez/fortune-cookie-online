FROM golang:latest

RUN apt-get update -y

WORKDIR /app
COPY . .

ENTRYPOINT ["go", "run", "."]