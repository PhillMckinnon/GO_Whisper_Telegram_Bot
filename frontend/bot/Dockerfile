FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o telegram-bot main.go

CMD ["./telegram-bot"]