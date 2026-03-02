FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /twix ./cmd/api

FROM alpine:3.23.3

RUN apk add --no-cache ca-certificates

COPY --from=builder /twix /twix

EXPOSE 8080

CMD ["/twix"]
