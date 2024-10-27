# SpaceTrouble 🚀

SpaceTrouble is a Go-based REST API for booking space travel tickets. It enables users to create and manage bookings for space trips while ensuring launchpad availability and avoiding conflicts with SpaceX launches.

## Features 🌟

- Create and view space travel bookings
- Real-time integration with SpaceX API to check launch conflicts
- Validation of booking requests (age, destination, launchpad availability)
- Support for multiple destinations (Mars, Moon, Pluto, etc.)
- PostgreSQL database for persistent storage
- JSON and XML response formats
- Cursor-based pagination for booking listings
- Health check endpoint with system metrics
- Docker containerization for easy deployment

## Prerequisites 📋

- Docker and Docker Compose
- Go 1.22 or higher (for local development)
- Make (optional, but recommended)
- PostgreSQL 16 (handled by Docker)

## Quick Start 🚀

1. Clone the repository:
```bash
git clone https://github.com/chrisdamba/spacetrouble.git
cd spacetrouble
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Build and start the containers:
```bash
make docker-build
make docker-up
```

4. Run database migrations:
```bash
make migrate-up
```

The API will be available at `http://localhost:5000`

## API Endpoints 🛠️

### Create Booking
```http
POST /v1/bookings
Content-Type: application/json

{
    "first_name": "John",
    "last_name": "Doe",
    "gender": "male",
    "birthday": "1990-01-01T00:00:00Z",
    "launchpad_id": "5e9e4502f5090995de566f86",
    "destination_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
    "launch_date": "2025-01-01T00:00:00Z"
}
```

### Get Bookings
```http
GET /v1/bookings
Accept: application/json
```

### Health Check
```http
GET /v1/health
```

## Environment Variables ⚙️

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_ADDRESS | Server listening address | :5000 |
| SERVER_READ_TIMEOUT | HTTP read timeout | 15s |
| SERVER_WRITE_TIMEOUT | HTTP write timeout | 15s |
| SERVER_IDLE_TIMEOUT | HTTP idle timeout | 30s |
| POSTGRES_HOST | PostgreSQL host | localhost |
| POSTGRES_PORT | PostgreSQL port | 5432 |
| POSTGRES_DB | PostgreSQL database name | space |
| POSTGRES_USER | PostgreSQL username | postgres |
| POSTGRES_PASSWORD | PostgreSQL password | postgres |
| MAX_CONNS | Max DB connections | 99 |
| SPACEX_URL | SpaceX API base URL | https://api.spacexdata.com/v4 |

## Project Structure 📁

```
spacetrouble/
├── cmd/
│   └── api/ 
│       ├────main.go            # Application entry point
├── internal/
│   ├── api/                    # API handlers
│   ├── models/                 # Domain models
│   ├── repository/             # Database operations
│   ├── service/                # Business logic
│   └── validator/              # Request validation
├── pkg/
│   ├── config/                 # Configuration management
│   ├── health/                 # Health check endpoint
│   └── spacex/                 # SpaceX API client
├── migrations/                 # Database migrations
├── Dockerfile                  # Docker build instructions
├── docker-compose.yml         # Docker compose configuration
├── Makefile                   # Build and development commands
└── README.md                  # Project documentation
```

## Available Make Commands 🛠️

- `make build`: Build the application
- `make test`: Run tests
- `make vet`: Run Go vet
- `make docker-build`: Build Docker image
- `make docker-up`: Start Docker containers
- `make docker-down`: Stop Docker containers
- `make migrate-up`: Run database migrations
- `make migrate-down`: Revert database migrations

## Testing 🧪

Run the test suite:
```bash
make test
```

## Deployment 🚢

The application is containerised and can be deployed using Docker:

1. Build the image:
```bash
make docker-build
```

2. Configure environment variables for your deployment environment

3. Run the containers:
```bash
make docker-up
```

## Technical Details 🔧

- **Architecture**: Clean Architecture pattern
- **API Design**: RESTful with JSON/XML support
- **Database**: PostgreSQL with migrations
- **External Integration**: SpaceX API for launch checks
- **Validation**: Custom validation rules for bookings
- **Error Handling**: Structured error responses
- **Monitoring**: Health check endpoint with metrics
