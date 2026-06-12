# User Age API

A production-ready REST API built with GoFiber, PostgreSQL, SQLC, Uber Zap, and go-playground/validator.

## Architecture

Clean Architecture layers:

```
handler → service → repository → SQLC → PostgreSQL
```

- **Handler** — HTTP layer, request parsing, validation
- **Service** — Business logic (age calculation)
- **Repository** — Data access via SQLC generated code
- **SQLC** — Type-safe SQL query generation

## Endpoints

| Method | Path          | Description     |
|--------|---------------|-----------------|
| POST   | /api/users    | Create a user   |
| GET    | /api/users    | List all users  |
| GET    | /api/users/:id| Get user by ID  |
| PUT    | /api/users/:id| Update a user   |
| DELETE | /api/users/:id| Delete a user   |

## Request / Response

**POST /api/users**
```json
{ "name": "Alice", "dob": "1990-05-15" }
```
```json
{ "id": 1, "name": "Alice", "dob": "1990-05-15", "age": 36 }
```

**GET /api/users**
```json
[
  { "id": 1, "name": "Alice", "dob": "1990-05-15", "age": 36 }
]
```

**PUT /api/users/1**
```json
{ "name": "Alice Updated", "dob": "1991-06-20" }
```

**DELETE /api/users/1** returns `204 No Content`.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose

### Run with Docker

```bash
docker compose up --build
```

### Run locally

1. Start PostgreSQL and create the `userage` database.
2. Run the migration:

```bash
psql -U postgres -d userage -f db/migrations/001_create_users_table.sql
```

3. Start the server:

```bash
go run ./cmd/server
```

### Environment variables

| Variable      | Default                                                         |
|---------------|-----------------------------------------------------------------|
| SERVER_PORT   | 8080                                                            |
| DATABASE_URL  | postgres://postgres:postgres@localhost:5432/userage?sslmode=disable |
| READ_TIMEOUT  | 10                                                              |
| WRITE_TIMEOUT | 10                                                              |

## Project Structure

```
├── cmd/server/main.go          # Entry point, dependency injection
├── config/config.go            # Configuration from env vars
├── db/
│   ├── migrations/             # SQL migrations
│   ├── query/                  # SQLC query definitions
│   └── sqlc/                   # Generated SQLC code
├── internal/
│   ├── handler/                # HTTP handlers
│   ├── service/                # Business logic
│   ├── repository/             # Data access layer
│   ├── routes/                 # Route registration
│   ├── middleware/              # Request ID, logging, error handler
│   ├── models/                 # Data transfer objects
│   └── logger/                 # Zap logger setup
├── sqlc.yaml                   # SQLC configuration
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Age Calculation

Age is computed dynamically from the date of birth stored in the database using:

```go
func calculateAge(dob time.Time) int {
    now := time.Now()
    age := now.Year() - dob.Year()
    if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
        age--
    }
    return age
}
```
