# Build stage
FROM golang:1.25-alpine AS builder

# Note: This Dockerfile must be built from the repository root:
# docker build -f go/kaws/Dockerfile -t kaws-operator:latest .

WORKDIR /workspace

# Copy monorepo structure to handle replace directives
COPY go.mod go.sum ./
COPY pkg ./pkg
COPY go/kaws ./go/kaws

WORKDIR /workspace/go/kaws

# Download dependencies (replace directive now works)
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o kaws main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy binary from builder
COPY --from=builder /workspace/go/kaws/kaws .

USER 65532:65532

ENTRYPOINT ["/kaws"]

