# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
# COPY go.sum ./ # No go.sum yet if no external deps beyond stdlib or untidy.
# But we added prometheus deps, so we should run go mod tidy first.
# Assuming we will run go mod tidy before build or here.
RUN go mod download

# Copy source code
COPY . .

# Ensure deps are tidy (needed because local go version might be old)
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /lotw-exporter ./cmd/lotw-exporter

# Final Stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /lotw-exporter /lotw-exporter

EXPOSE 8080

ENTRYPOINT ["/lotw-exporter"]
