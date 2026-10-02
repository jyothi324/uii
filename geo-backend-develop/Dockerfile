# Build stage
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o api ./cmd

# Run stage - using distroless for minimal size
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/api .

EXPOSE 3000
ENTRYPOINT ["./api"]