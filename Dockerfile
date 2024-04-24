FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o main cmd/main/main.go

EXPOSE 8080

CMD ["./main"]