# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/proxi-pulse ./cmd/proxi-pulse

# Runtime stage
FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /app/proxi-pulse /usr/local/bin/proxi-pulse
COPY --from=builder /app/web /web
ENTRYPOINT ["proxi-pulse"]
