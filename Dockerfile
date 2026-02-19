FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./app/main.go


FROM alpine:3.19
WORKDIR /app


RUN apk add --no-cache ca-certificates tzdata postgresql-client && update-ca-certificates

COPY --from=builder /app/main .
COPY migrations ./migrations
COPY docs ./docs
COPY scripts/wait-for-postgres.sh /usr/local/bin/wait-for-postgres.sh
RUN chmod +x /usr/local/bin/wait-for-postgres.sh

EXPOSE 8080
CMD ["wait-for-postgres.sh", "./main"]

