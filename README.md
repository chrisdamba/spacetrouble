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

## Running the Project 🚀

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/chrisdamba/spacetrouble.git
cd spacetrouble
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your preferred settings
```

3. Start the services:
```bash
# Build the Docker images
make docker-build

# Start all services
make docker-up

# Run database migrations
make migrate-up
```

4. Verify the setup:
```bash
# Check if services are running
docker-compose ps

# Check application logs
docker-compose logs -f app

# Test the health endpoint
curl http://localhost:5000/v1/health
```

### Stopping the Project
```bash
# Stop all services
make docker-down

# Stop and remove all containers, networks, and volumes
docker-compose down -v
```

### Rebuilding After Changes
```bash
# Rebuild the application
make docker-build

# Restart services
make docker-down
make docker-up
```

The API will be available at `http://localhost:5000`
I'll update the API Endpoints section of the README.md with the new endpoints and more detailed information:

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
Response (201 Created):
```json
{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user": {
        "id": "123e4567-e89b-12d3-a456-426614174001",
        "first_name": "John",
        "last_name": "Doe",
        "gender": "male",
        "birthday": "1990-01-01T00:00:00Z"
    },
    "flight": {
        "id": "123e4567-e89b-12d3-a456-426614174002",
        "launchpad_id": "5e9e4502f5090995de566f86",
        "destination": {
            "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
            "name": "Mars"
        },
        "launch_date": "2025-01-01T00:00:00Z"
    },
    "status": "ACTIVE",
    "created_at": "2024-01-01T00:00:00Z"
}
```

### List Bookings
```http
GET /v1/bookings?limit=10&cursor=<cursor_token>
Accept: application/json
```
Response (200 OK):
```json
{
    "bookings": [
        {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "user": {
                "id": "123e4567-e89b-12d3-a456-426614174001",
                "first_name": "John",
                "last_name": "Doe",
                "gender": "male",
                "birthday": "1990-01-01T00:00:00Z"
            },
            "flight": {
                "id": "123e4567-e89b-12d3-a456-426614174002",
                "launchpad_id": "5e9e4502f5090995de566f86",
                "destination": {
                    "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
                    "name": "Mars"
                },
                "launch_date": "2025-01-01T00:00:00Z"
            },
            "status": "ACTIVE",
            "created_at": "2024-01-01T00:00:00Z"
        }
    ],
    "limit": 10,
    "cursor": "next_page_token"
}
```

### Delete Booking
```http
DELETE /v1/bookings?id=123e4567-e89b-12d3-a456-426614174000
```
Response (204 No Content)

### Health Check
```http
GET /v1/health
```
Response (200 OK):
```json
{
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "uptime": "24h0m0s",
    "go_version": "go1.22",
    "memory": {
        "alloc": 1234567,
        "totalAlloc": 2345678,
        "sys": 3456789,
        "numGC": 10
    }
}
```

### Available Destinations
| Destination | ID |
|-------------|------|
| Mars | a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11 |
| Moon | b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22 |
| Pluto | c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33 |
| Asteroid Belt | d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44 |
| Europa | e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55 |
| Titan | f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66 |
| Ganymede | 70eebc99-9c0b-4ef8-bb6d-6bb9bd380a77 |

### Error Responses
| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid input data |
| 404 | Not Found - Booking or destination not found |
| 409 | Conflict - Launchpad unavailable or SpaceX conflict |
| 500 | Internal Server Error |

Example error response:
```json
{
    "error": "launchpad is unavailable"
}
```

### Request Validation Rules
- `first_name`, `last_name`: Required, max 50 characters
- `gender`: Must be "male", "female", or "other"
- `birthday`: Must be between 18-75 years old
- `launchpad_id`: Must be 24 characters
- `destination_id`: Must be a valid UUID from available destinations
- `launch_date`: Must be in the future


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
├── tests/                      # Tests
│   ├── api/
│   ├── mocks/
│   ├── pkg/
│   ├── repository/
│   ├── service/
│   ├── utils/
│   └── validator/
├── migrations/                 # Database migrations
├── Dockerfile                  # Docker build instructions
├── docker-compose.yml          # Docker compose configuration
├── Makefile                    # Build and development commands
└── README.md                   # Project documentation
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

## Testing Structure 🧪

### Test Organization
```
tests/
├── api/                            # API handler tests
│   └── api_test.go                 # Tests for API endpoints
├── mocks/                          # Mock implementations
│   ├── booking_repository.go       # Mock repository
│   └── spacex_client.go            # Mock SpaceX client
├── pkg/                            # Package tests
│   ├── config/                     # Configuration tests
│   ├── health/                     # Health check tests
│   └── spacex/                     # SpaceX client tests
├── repository/                     # Repository layer tests
│   └── booking_repository_test.go
├── service/                        # Service layer tests
│   └── booking_service_test.go
├── utils/                          # Test utilities
│   └── mock_data.go                # Test data generators
└── validator/                      # Validation tests
└── validator_test.go
```

### Test Coverage

#### API Tests (`tests/api/`)
- Tests HTTP endpoints
- Validates request/response formats
- Checks content type handling
- Verifies error responses
- Tests pagination
- Validates HTTP method restrictions

#### Repository Tests (`tests/repository/`)
- Database CRUD operations
- Transaction handling
- Error scenarios
- Connection pooling
- Query building
- Cursor-based pagination

#### Service Tests (`tests/service/`)
- Business logic validation
- SpaceX integration
- Booking workflow
- Error handling
- Data transformation
- Business rules enforcement

#### Mock Implementations (`tests/mocks/`)
- `MockBookingRepository`: Repository layer mocking
  - Simulates database operations
  - Provides predictable responses
  - Enables error scenario testing

- `MockSpaceXClient`: SpaceX API mocking
  - Simulates API responses
  - Tests network failures
  - Validates integration points

#### Test Utilities (`tests/utils/`)
- `mock_data.go`: Test data generation
  - Creates consistent test bookings
  - Generates valid UUIDs
  - Provides sample requests/responses
  - Helps compare complex objects

### Running Tests

#### All Tests
```bash
make test
```

#### Specific Package Tests
```bash
go test ./tests/service/...
go test ./tests/repository/...
go test ./tests/api/...
```

#### With Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Categories

#### Unit Tests
- Individual component testing
- Mock external dependencies
- Fast execution
- High coverage

#### Integration Tests
- Database interaction
- API endpoint behavior
- External service integration
- Real-world scenarios

#### Mock Tests
Tests using mock implementations to verify:
- Error handling
- Edge cases
- Resource unavailability
- Invalid data scenarios

### Test Scenarios

#### Booking Creation Tests
- Valid booking creation
- Invalid input validation
- Destination verification
- Launch date conflicts
- SpaceX availability
- Database errors
- Transaction rollback

#### Booking Listing Tests
- Pagination functionality
- Cursor handling
- Empty results
- Large result sets
- Filter application
- Sort ordering

#### Booking Deletion Tests
- Successful deletion
- Non-existent booking
- Invalid UUID
- Unauthorized deletion
- Status validation

#### Error Handling Tests
- Database connection failures
- SpaceX API unavailability
- Invalid input data
- Business rule violations
- Concurrent access

### Test Design Principles

1. **Isolation**
   - Tests run independently
   - No shared state
   - Clean setup/teardown

2. **Reproducibility**
   - Consistent test data
   - Deterministic results
   - Clear failure messages

3. **Coverage**
   - Happy path scenarios
   - Error conditions
   - Edge cases
   - Business rules

4. **Maintainability**
   - Clear test names
   - Shared utilities
   - Common patterns
   - Documentation

### Writing New Tests

1. Create test file in appropriate directory
```go
package repository_test

func TestNewFeature(t *testing.T) {
    t.Run("successful case", func(t *testing.T) {
        // Test setup
        // Test execution
        // Assertions
    })

    t.Run("error case", func(t *testing.T) {
        // Test setup
        // Test execution
        // Assertions
    })
}
```

2. Use test utilities
```go
booking := utils.CreateMockBooking(uuid.Nil)
bookings := utils.CreateMockBookings(5)
```

3. Use mocks
```go
mockRepo := new(mocks.MockBookingRepository)
mockSpaceX := new(mocks.MockSpaceXClient)
```

4. Assert expectations
```go
assert.NoError(t, err)
assert.Equal(t, expected, actual)
mockRepo.AssertExpectations(t)
```

### Common Test Patterns

1. **Table-Driven Tests**
```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {"valid case", "input", "expected", false},
    {"error case", "bad input", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

2. **Setup/Teardown**
```go
func setupTest(t *testing.T) (*mocks.MockBookingRepository, *service.BookingService) {
    mockRepo := new(mocks.MockBookingRepository)
    svc := service.NewBookingService(mockRepo)
    return mockRepo, svc
}
```

3. **Helper Functions**
```go
func assertBookingsEqual(t *testing.T, expected, actual *models.Booking) {
    t.Helper()
    assert.Equal(t, expected.ID, actual.ID)
    // Additional assertions
}
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


## Troubleshooting 🔧

### Common Docker Issues

1. **Port Conflicts**
```bash
Error starting userland proxy: listen tcp 0.0.0.0:5000: bind: address already in use
```
Solution:
```bash
# Find the process using the port
sudo lsof -i :5000
# Kill the process
kill -9 <PID>
# Or change the port in docker-compose.yml
```

2. **Database Connection Issues**
```bash
error: connect: connection refused
```
Solutions:
- Check if the database container is running:
```bash
docker-compose ps
```
- Verify database credentials in `.env`
- Try restarting the services:
```bash
make docker-down
make docker-up
```

3. **Permission Issues**
```bash
permission denied while trying to connect to the Docker daemon socket
```
Solution:
```bash
# Add your user to the docker group
sudo usermod -aG docker ${USER}
# Log out and back in or run:
newgrp docker
```

4. **Database Migration Failures**
```bash
error: migration failed
```
Solutions:
- Check migration logs:
```bash
docker-compose logs app
```
- Reset migrations:
```bash
make migrate-down
make migrate-up
```
- Verify database connection:
```bash
docker-compose exec db psql -U postgres -d space -c "\l"
```

5. **Container Not Starting**
```bash
Error response from daemon: Container is not running
```
Solutions:
- Check container logs:
```bash
docker-compose logs <service_name>
```
- Verify container status:
```bash
docker-compose ps
```
- Remove containers and volumes:
```bash
docker-compose down -v
docker-compose up -d
```

6. **Image Build Failures**
```bash
failed to build: exit status 1
```
Solutions:
- Clean Docker cache:
```bash
docker system prune -a
```
- Rebuild with no cache:
```bash
docker-compose build --no-cache
```

### Maintenance Commands

1. **Reset Everything**
```bash
# Stop all containers and remove volumes
make docker-down
docker-compose down -v

# Remove all related images
docker rmi $(docker images | grep spacetrouble)

# Rebuild from scratch
make docker-build
make docker-up
make migrate-up
```

2. **View Logs**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f db
```

3. **Access Database**
```bash
# Connect to PostgreSQL
docker-compose exec db psql -U postgres -d space

# Backup database
docker-compose exec db pg_dump -U postgres space > backup.sql

# Restore database
cat backup.sql | docker-compose exec -T db psql -U postgres -d space
```

4. **Check Container Status**
```bash
# List containers
docker-compose ps

# Container details
docker inspect <container_id>

# Resource usage
docker stats
```

### Performance Tuning

If you experience performance issues:

1. **Database Tuning**
    - Adjust `max_connections` in PostgreSQL
    - Modify connection pool size in `.env`
    - Monitor query performance with:
```bash
docker-compose exec db psql -U postgres -d space -c "SELECT * FROM pg_stat_activity;"
```

2. **Application Tuning**
    - Adjust timeouts in `.env`
    - Monitor memory usage:
```bash
docker stats spacetrouble_app_1
```

3. **System Resources**
    - Increase Docker resources (CPU/Memory) in Docker Desktop settings
    - Monitor resource usage:
```bash
docker-compose top
```

## Technical Details 🔧

- **Architecture**: Clean Architecture pattern
- **API Design**: RESTful with JSON/XML support
- **Database**: PostgreSQL with migrations
- **External Integration**: SpaceX API for launch checks
- **Validation**: Custom validation rules for bookings
- **Error Handling**: Structured error responses
- **Monitoring**: Health check endpoint with metrics
