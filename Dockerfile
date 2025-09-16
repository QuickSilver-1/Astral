FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w -extldflags "-static"' -a -o main ./cmd/astral

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

ENTRYPOINT ["./main"]