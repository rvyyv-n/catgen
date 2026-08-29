# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cats .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/cats .
COPY --from=builder /app/images ./images
COPY --from=builder /app/favicon.svg .

EXPOSE 8090

ENV PORT=8090
ENV IMAGES_DIR=images

CMD ["./cats"]