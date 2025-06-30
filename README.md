# Backend Services Project

Multi-service backend with nginx reverse proxy and PostgreSQL database.

## Services

- **Nginx**: Reverse proxy (port 80)
- **Go Service**: REST API with database
- **PostgreSQL**: Database (port 5432)

## Quick Start

```bash
# Start all services
docker compose up --build

# Test endpoints
curl http://localhost/health

# Upload a JSON file
curl -X POST http://localhost/api/v1/files \
  -F "name=myfile.json" \
  -F "file=@./myfile.json"

# List all uploaded files
curl http://localhost/api/v1/files

# Download a file by id (replace 1 with actual id)
curl -OJ http://localhost/api/v1/files/1
```

## API Endpoints

- `GET /health` - Nginx health check
- `POST /api/v1/files` - Upload a JSON file (multipart form: `name`, `file`)
- `GET /api/v1/files` - List all uploaded files
- `GET /api/v1/files/{id}` - Download a file by id

## Stop Services

```bash
docker compose down
``` 