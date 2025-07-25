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

## Running the service
```makefile
# Build the application
make all

# Run tests
make test

# Run linting
make lint

# Clean build artifacts
make clean

```
## Using docker-compose
```bash
# Start all services
docker-compose up

# Stop all services
docker-compose down
```
The docker-compose.yml file includes:
- redis: Used for caching shortened URLs 
- postgres: Database for persistent storage
