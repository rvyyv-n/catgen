FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cats .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/cats .
COPY --from=builder /app/images ./images
COPY --from=builder /app/favicon.svg .

EXPOSE 8090

ENV PORT=8090
ENV IMAGES_DIR=images

CMD ["./cats", "--server"]