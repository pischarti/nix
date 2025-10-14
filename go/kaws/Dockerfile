# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o kaws main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy binary from builder
COPY --from=builder /workspace/kaws .

USER 65532:65532

ENTRYPOINT ["/kaws"]

