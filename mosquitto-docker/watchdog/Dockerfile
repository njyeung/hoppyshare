# Dockerfile
FROM golang:1.21

WORKDIR /app

COPY go.mod ./
COPY main.go ./
COPY tls.go ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o watchdog .

CMD ["./watchdog"]

