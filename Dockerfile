FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main/main.go 

EXPOSE 8080

FROM postgres:latest AS postgres

COPY init.sql /docker-entrypoint-initdb.d/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .  
RUN chmod +x main
COPY --from=postgres /usr/local/bin/docker-entrypoint.sh /usr/local/bin/

CMD ["./main"]
