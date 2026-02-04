# ---------- Build stage ----------
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .

# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

# Copy binary
COPY --from=builder /app/myapp /app/myapp

# Copy templates
COPY --from=builder /app/templates /app/templates

# Create uploads dir
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["/app/myapp"]