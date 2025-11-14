# PR Reviewer Assignment Service

This project implements a service for automatically assigning reviewers to Pull Requests (PRs), managing teams, and users, as specified in the Backend Trainee Assignment (Autumn 2025).

## Features

-   **Team Management:** Create and retrieve teams with their members.
-   **User Management:** Set user activity status.
-   **Pull Request Management:**
    -   Create PRs with automatic assignment of up to two active reviewers from the author's team (excluding the author).
    -   Merge PRs (idempotent operation).
    -   Reassign reviewers (replaces an existing reviewer with a random active member from their team).
    -   Retrieve PRs assigned to a specific user.
-   **Database:** PostgreSQL for data persistence.
-   **API:** Fully compliant with the provided `openapi.yml` specification.
-   **Containerization:** Dockerized application and database for easy deployment.

## Project Structure

The project follows a clean architecture pattern, separating concerns into distinct layers:

-   `api/`: Contains the `openapi.yml` specification and generated Go code (models, server interfaces) using `oapi-codegen`.
-   `cmd/`: Application entry points.
    -   `cmd/main.go`: Main application entry point, responsible for bootstrapping the service.
-   `configs/`: Application configuration loading from environment variables.
-   `db/`: Database connection utilities.
-   `migrations/`: SQL migration files for database schema management.
-   `pkg/utils/`: Utility functions, such as the structured logger.
-   `services/`: Contains the core business logic (service layer) and data access implementations (repository layer), grouped by domain.
    -   `services/team/`: Team-related business logic and repository.
    -   `services/user/`: User-related business logic and repository.
    -   `services/pullrequest/`: Pull Request-related business logic and repository.
-   `types/`: Defines Go interfaces for repositories and services, acting as contracts between layers.
-   `internal/handler/`: HTTP handlers that implement the API interfaces and interact with the service layer.

## Technologies Used

-   **Go:** Programming language.
-   **PostgreSQL:** Database.
-   **Docker & Docker Compose:** Containerization and orchestration.
-   **`oapi-codegen`:** OpenAPI specification code generation.
-   **`go-chi/chi`:** Lightweight HTTP router.
-   **`jackc/pgx/v5`:** PostgreSQL driver.
-   **`golang-migrate/migrate`:** Database migration tool.
-   **`joho/godotenv`:** Environment variable loading from `.env` files.
-   **`log/slog`:** Structured logging.

## Setup and Running

### Prerequisites

-   Go (1.25+)
-   Docker & Docker Compose
-   `oapi-codegen` (installed via `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`)
-   `golangci-lint` (for linting, optional but recommended)

### Environment Variables

The application uses environment variables for configuration. You can create a `.env` file in the project root:

```
HTTP_PORT=8080
DATABASE_URL=postgres://user:password@db:5432/avito_trainee?sslmode=disable
ENV=development
```

### Commands

Use the provided `Makefile` for common operations:

-   **`make build`**: Builds the Go application binary.
-   **`make run`**: Builds and runs the application locally (requires a running PostgreSQL instance or `make docker-up`).
-   **`make docker-up`**: Builds the Docker images and starts the application and PostgreSQL database using Docker Compose. This also applies database migrations automatically.
-   **`make docker-down`**: Stops and removes the Docker Compose environment.
-   **`make migrate-up`**: Applies database migrations to the configured `DATABASE_URL`.
-   **`make migrate-down`**: Rolls back database migrations.
-   **`make test`**: Runs unit tests.
-   **`make test-integration`**: Runs integration tests. This command automatically starts a separate Docker Compose environment with a test database (`avito_trainee_test`), runs tests, and then tears down the environment.
-   **`make lint`**: Runs `golangci-lint` to check code quality.
-   **`make clean`**: Removes build artifacts.

### Running the Service

1.  **Start the Docker Compose environment:**
    ```bash
    make docker-up
    ```
    This will build the application image, start PostgreSQL, and apply migrations.
2.  **Access the API:**
    The service will be available at `http://localhost:8080`. You can use tools like `curl` or Postman to interact with the API endpoints defined in `openapi.yml`.

    **Example:**
    ```bash
    curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d '{
        "team_name": "backend",
        "members": [
            {"user_id": "u1", "username": "Alice", "is_active": true},
            {"user_id": "u2", "username": "Bob", "is_active": true}
        ]
    }'
    ```

## Testing

-   **Unit Tests:**
    ```bash
    make test
    ```
-   **Integration Tests:**
    ```bash
    make test-integration
    ```
    This will spin up a dedicated test database, run the tests against it, and then clean up.

## Additional Notes

-   **Error Handling:** The service uses structured error responses as defined in `openapi.yml`. Specific error codes are returned for various scenarios (e.g., `NOT_FOUND`, `TEAM_EXISTS`).
-   **Idempotency:** The `POST /pullRequest/merge` endpoint is idempotent. Repeated calls will not cause errors and will return the current state of the PR.
-   **Assumptions:**
    -   User IDs and Pull Request IDs are unique strings.
    -   Team names are unique.
    -   Reviewer assignment is random among active team members (excluding the author).
    -   The `pg_isready` command is available in the `db` service for health checks.

## Future Improvements (Beyond the Scope of this Assignment)

-   Implement custom error types in the service layer and map them explicitly to `api.ErrorResponseErrorCode` in the handler.
-   Add more robust input validation beyond what `oapi-codegen` provides.
-   Implement a more sophisticated logging strategy (e.g., request IDs, context-aware logging).
-   Add authentication and authorization.
-   Implement the "Additional Tasks" from the assignment brief (statistics, load testing, bulk deactivation).
