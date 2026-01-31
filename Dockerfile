# Stage 1: Build the Go binary
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp ./main.go

# Stage 2: Create small runtime image
FROM alpine:latest

WORKDIR /

COPY --from=builder /myapp /myapp

# If your app needs ca-certificates or tzdata, add them:
# RUN apk --no-cache add ca-certificates

CMD ["/myapp"]