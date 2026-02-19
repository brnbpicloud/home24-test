# URL Tracker & Analyzer

A distributed web application for tracking and analyzing URLs with comprehensive page metrics. Built with Go, Redis, and Docker.

## Features

- **URL Submission & Tracking**: Submit URLs for analysis via web interface
- **Background Processing**: Redis-backed job queue with dedicated worker service
- **Comprehensive Analysis**: Extracts page title, HTML version, heading counts, link categorization, and login form detection
- **Real-time Dashboard**: View all tracked URLs with sortable/filterable DataTables interface
- **Containerized Deployment**: Complete Docker Compose setup for easy deployment

## Architecture

- **Web Service** (Port 4000): Frontend server with Bootstrap UI and API proxy
- **API Service** (Port 4001): Backend REST API for URL management
- **Worker Service**: Background processor that polls Redis queue and crawls URLs
- **Redis** (Port 6379): Job queue and data persistence with AOF enabled

## Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development only)

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone git@github.com:brnbpicloud/home24-test.git
   cd home24-test
   ```

2. **Start all services**

   ```bash
   docker-compose up --build
   ```

3. **Access the application**
   - Web Interface: http://localhost:4000
   - API Endpoint: http://localhost:4001

4. **Stop services**
   ```bash
   docker-compose down
   ```

### Local Development

1. **Start Redis**

   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   ```

2. **Run API server**

   ```bash
   REDIS_ADDR=localhost:6379 SERVER_PORT=4001 go run ./api
   ```

3. **Run Web server**

   ```bash
   API_ADDR=http://localhost:4001 REDIS_ADDR=localhost:6379 SERVER_PORT=4000 go run ./web
   ```

4. **Run Worker**
   ```bash
   REDIS_ADDR=localhost:6379 go run ./worker
   ```

## Usage

1. Navigate to http://localhost:4000
2. Enter a URL in the input field (e.g., `https://google.com`)
3. Click "Analyze" to submit for analysis
4. View the tracking list with real-time status updates
5. Click "View Details" to see comprehensive analysis results

## Analysis Metrics

The crawler extracts the following information:

- **HTML Version**: Detected HTML doctype
- **Page Title**: Title tag content
- **Heading Counts**: Count of H1-H6 elements
- **Links**:
  - Internal links (same domain)
  - External links (different domains)
  - Inaccessible links (broken/404s)
- **Login Form**: Detects presence of login input fields
- **HTTP Status Code**: Response status from crawl

## Environment Variables

### API Service

- `REDIS_ADDR`: Redis connection address (default: `localhost:6379`)
- `SERVER_PORT`: API server port (default: `4001`)

### Web Service

- `API_ADDR`: Backend API URL (default: `http://localhost:4001`)
- `REDIS_ADDR`: Redis connection address (default: `localhost:6379`)
- `SERVER_PORT`: Web server port (default: `4000`)

### Worker Service

- `REDIS_ADDR`: Redis connection address (default: `localhost:6379`)

## Tests

Run tests for each service:

```bash
# API tests
go test ./api/...

# Web tests
go test ./web/...

# Worker tests
go test ./worker/...
```

Run all tests:

```bash
go test ./...
```

## Docker Commands

```bash
# Build and start all services
docker-compose up --build

# Run in detached mode
docker-compose up -d --build

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f web
docker-compose logs -f api
docker-compose logs -f worker

# Rebuild single service
docker-compose up -d --build web

# Remove volumes (clears Redis data)
docker-compose down -v
```

## Dependencies

- **Chi**: Lightweight HTTP router
- **Colly**: Web scraping framework
- **go-redis**: Redis client for Go
- **UUID**: Unique identifier generation
- **Bootstrap 5**: Frontend CSS framework
- **DataTables**: Interactive table plugin

## Data Persistence

Redis data is persisted to a Docker volume (`redis_data`).

## Development Notes

- Status transitions: `pending` → `processing` → `completed` or `failed`
- Potential future improvements:
  - Use Kubernetes to orchestrate and scale workers
  - Adopt an event-driven architecture for communication between frontend, backend, and workers
  - Ensure unique entries per site
  - Add the ability to retry analysis (via a UI button or scheduled job)
  - Provide email notification reports, especially for failed sites
