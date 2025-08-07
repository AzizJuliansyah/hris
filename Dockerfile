# Dockerfile
FROM golang:1.24.4

# Install Air
RUN go install github.com/air-verse/air@v1.62.0


WORKDIR /app

# Copy all code
COPY . .

# Download dependencies
RUN go mod tidy

EXPOSE 8000

CMD ["air"]
