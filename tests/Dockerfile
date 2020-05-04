FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -a -ldflags="-s" -o /app/server /app/server.go

CMD ["/app/server"]