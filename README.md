# Short URL Service
This project implements a URL shortening service that allows users to convert long URLs into shorter, more manageable links.


## Project Description
The Short URL Service provides an API to:
- Create shortened URLs from long URLs
- Redirect users from shortened URLs to their original destinations
- Track URL usage statistics

### Prerequisites
Go 1.24
Docker and Docker Compose
Make

### Project structure
├── cmd/            # Application entrypoints
├── internal/       # Private application code
├── pkg/            # Public libraries
├── Makefile        # Build automation
└──docker-compose.yml # Container orchestration


## Running the service
```makefile
# Build the application
make build

# Run the service locally
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linting
make lint

# Clean build artifacts
make clean

# Build Docker image
make docker-build

# Run the service in Docker
make docker-run
```
## Using docker-compose
```bash
# Start all services
docker-compose up

# Run in detached mode
docker-compose up -d

# Stop all services
docker-compose down
```
The docker-compose.yml file includes:
- redis: Used for caching shortened URLs 
- postgres: Database for persistent storage
