# Build React app stage
FROM node:18-alpine as react-builder

# Set working directory for React app
WORKDIR /app/frontend

# Copy React app package files
COPY chat_app/package*.json ./

# Install dependencies
RUN npm install

# Copy React app source code
COPY chat_app/ ./

# Build React app
RUN npm run build

# Build Go app stage
FROM golang:1.23.3-alpine AS go-builder

# Add git and build dependencies for SQLite
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:3.19

# Add ca-certificates and SQLite runtime dependencies
RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /root/

# Copy the binary from builder
COPY --from=go-builder /app/main .

# Copy the built React app
COPY --from=react-builder /app/frontend/dist /root/static

# Create a directory for the database
RUN mkdir -p /data
VOLUME /data

# Environment variable for database path
ENV DB_PATH=/data/turplemq.db

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]