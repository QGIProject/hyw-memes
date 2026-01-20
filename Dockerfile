# --- Frontend Build Stage ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# --- Backend Build Stage ---
FROM golang:alpine AS backend-builder
WORKDIR /app
# Install build dependencies for CGO if needed (though we aim for CGO_ENABLED=0)
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# --- Final Production Stage ---
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies (e.g., libwebp for utilities if needed)
RUN apk add --no-cache libwebp-tools ca-certificates

# Copy the server binary
COPY --from=backend-builder /app/server .
# Copy the frontend build artifacts
COPY --from=frontend-builder /app/web/dist ./web/dist

# Create uploads directory
RUN mkdir -p uploads

# Expose the application port
EXPOSE 3000

# Set environment variables (defaults)
ENV PORT=3000
ENV DATABASE_PATH=./memes.db
ENV UPLOAD_DIR=./uploads

# Entry point
CMD ["./server"]
