FROM golang:1.24-alpine AS backend

# Install required dependencies for Go
RUN apk add --no-cache git

# Set working directory for backend
WORKDIR /app

# Copy all source files first
COPY . .

# Download Go dependencies
RUN go mod download

# Build the Go application
RUN go build -o polyglot-server .

# Node.js stage for the frontend
FROM node:14-alpine AS frontend

# Set working directory for frontend
WORKDIR /app/ui

# Copy package.json and install dependencies
COPY ui/package.json ./
RUN npm install

# Copy the rest of the frontend code
COPY ui/ ./

# Final stage - combine both
FROM alpine:3.14

# Install necessary runtime dependencies
RUN apk add --no-cache nodejs npm ca-certificates

# Create app directory
WORKDIR /app

# Copy the Go binary from the backend stage
COPY --from=backend /app/polyglot-server .
COPY --from=backend /app/artifacts/ ./artifacts/

# Copy the frontend from the frontend stage
COPY --from=frontend /app/ui /app/ui

# Create startup script with proper line endings
RUN printf '#!/bin/sh\n\
# Start the backend server in the background\n\
echo "Starting Go backend server..."\n\
/app/polyglot-server &\n\
\n\
# Give the backend a moment to start\n\
sleep 2\n\
\n\
# Start the frontend server\n\
echo "Starting Node.js frontend server..."\n\
cd /app/ui && npm start\n' > /app/start.sh

# Make sure the script is executable
RUN chmod +x /app/start.sh

# Expose ports (8080 for the API, 3000 for the UI)
EXPOSE 8080 3000

# Set the entrypoint
ENTRYPOINT ["/app/start.sh"]
