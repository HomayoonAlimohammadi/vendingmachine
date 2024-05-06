# first stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o vendingmachine .

# second stage
FROM alpine

WORKDIR /app

COPY --from=builder /app/vendingmachine .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

CMD ["./vendingmachine"]
