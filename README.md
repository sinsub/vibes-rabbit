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
curl http://localhost/health                    # Nginx health
curl http://localhost/api/v1/hello              # Hello endpoint
curl http://localhost/api/v1/messages           # Get messages
curl -X POST http://localhost/api/v1/messages \  # Create message
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from API"}'
```

## API Endpoints

- `GET /health` - Nginx health check
- `GET /api/v1/hello` - Returns `{"hello":"world"}`
- `GET /api/v1/messages` - Get all messages
- `POST /api/v1/messages` - Create new message

## Stop Services

```bash
docker compose down
``` 