# Fixup

## About

This project is a backend application built using a microservices architecture. The backend is composed of multiple independent microservices, each handling a specific domain within the system:

- **User Service:** Manages user-related CRUD operations and authentication processes, including user registration, login (JWT-based with access and refresh tokens), refresh token rotation, and profile picture management via S3 and CloudFront.
- **Catalog Service:** Handles the management of service offerings, allowing predefined services to be assigned to users based on the tasks they are capable of performing.
- **Order Service:** Manages the workflow of service orders. It allows customers to request services, and providers can submit offers with their proposed terms. The customer can then choose an offer to proceed with.
- **Chat Service:** Enables communication between customers and service providers for order-related discussions.

### Application Structure

The application follows Domain-Driven Design (DDD) principles and employs layered architecture patterns to maintain a clean separation of concerns:

1. **Domain-Driven Design (DDD):**
   - Uses domain components such as entities and value objects to encapsulate core business components.
   - Domain events are processed internally within each microservice, maintaining encapsulated business logic.

2. **Repository Pattern:**
   - Abstracts data access using repositories to interact with the underlying data sources (Postgres, Redis).
   - The repository layer provides methods to manage database connections and operations, returning domain components or model components as needed.
   
3. **Service Layer:**
   - Contains business logic and handles transactions across multiple repositories when required.
   - Processes errors returned from the repository layer, translating them into meaningful service-level errors.
   
4. **Handler/Controller Layer:**
   - Acts as the entry point for HTTP requests, performing data validation, mapping to domain components, and handling responses.
   - Manages error handling and logging while formatting the final response for clients.

### Infrastructure

The application uses various tools and services to ensure reliability, scalability, and observability:

1. **Deployment Environment:**
   - Will be deployed on AWS, just an EC2 instance running docker wile pulling project image from ECR
2. **Containerization and Orchestration:**
   - Each microservice is containerized using Docker, with a `docker-compose.yml` file for local orchestration.
   - Each microservice has its own Dockerfile.

3. **CI/CD Pipeline:**
   - GitHub Actions are used for automated testing on pull requests and deployment to AWS upon merging changes.

4. **Monitoring and Logging:**
   - Monitoring is handled with Grafana and Prometheus, including basic metrics at `/metrics` endpoint and custom exporter for exposing NGINX statistics, based on access.log files, in addition CAdvisor for container stats.
   - Logging is managed using the ELK stack (Elasticsearch, Logstash, Kibana), with Filebeat reading log files and forwarding them to Logstash.

### Testing

- Tests follow a table-driven approach using Testify and uber's gommock for generating mocks.
- Repository layer is tested with Testcontainers, adapted to real world environment as much as possible.
- Each service has its own `_test.go` file for unit tests.
- You can run tests using `make test` or just repository test `make test-repo`

### SQL

- Project uses SQLC for generating raw go code for sql database interaction and you can generate a repository component running `make sqlc db={database}` a db argument is a name of a folder in sql directory that make file will look for
- Project also support database migration using go-migrate package, to migrate databse up or down, run `make migrate-up db={database}` or `make migrate-down db={database}`

### Documentation

- Each microservice has its own Swagger documentation, which provides an API reference for the available endpoints. The documentation is served through a Swagger UI Docker container, a container has it's own swagger config file.
- You can generate and update the Swagger documentation using a `make swag-gen svc={service}` command from the Makefile.
- If new microservice is added, modified or removed, you should indicate that in makefile, by changing values of `SVCS` and run `make swag-config`, which will look for new micrsoervices, genreate docs, and place them in `SWAGGER_CONF` location

## Security

Fixup follows good practices of security, but 100% security cannot be assured.
Fixup is provided **"as is"** without any **warranty**. Use at your own risk.

## License

This project is licensed under the **Apache Software License 2.0**.

See [LICENSE](LICENSE) for more information.

## Acknowledgements
- [golang - go](https://github.com/golang/go) The Go programming language. (BSD-3-Clause license)
- [prometheus - clien_golang](https://github.com/prometheus/prometheus) Prometheus instrumentation library for Go applications. (Apache-2.0 license)
- [grafana](https://github.com/grafana/grafana) The open and composable observability and data visualization platform. (AGPL-3.0 license)
- [redis - go redis](https://github.com/redis/go-redis) Redis Go client. (BSD-2-Clause license)
- [jackc - pgx](https://github.com/jackc/pgx) PostgreSQL driver and toolkit for Go. (MIT license)
- [go-gomail - gomail](https://github.com/go-gomail/gomail) The best way to send emails in Go. (MIT license)
- [go-chi - chi](https://github.com/go-chi/chi) lightweight, idiomatic and composable router for building Go HTTP services. (MIT license)
- [go-playground - validator](https://github.com/go-playground/validator) 💯Go Struct and Field validation, including Cross Field, Cross Struct. (MIT license)
- [bwmarrin - snowflake](https://github.com/bwmarrin/snowflake) A simple to use Go package to generate or parse Twitter snowflake IDs. (BSD-2-Clause license)
- [aws - aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) AWS SDK for the Go programming language.  (Apache-2.0 license)
- [golang-jwt - jwt](https://github.com/golang-jwt/jwt) Go implementation of JSON Web Tokens (JWT). (MIT license)
- [uber - zap](https://github.com/uber-go/zap) Blazing fast, structured, leveled logging in Go. (MIT license)
- [uber - gomock](https://github.com/uber-go/mock) GoMock is a mocking framework for the Go programming language. (Apache-2.0 license)
- [testcontainers](https://github.com/testcontainers/testcontainers-go) Create and clean up container-based dependencies for automated integration/smoke tests. (MIT license)
- [joho - godotenv](https://github.com/joho/godotenv) A Go port of Ruby's dotenv library (Loads environment variables from .env files). (MIT license)
- [go-swagger - go-swagger](https://github.com/go-swagger/go-swagger) Swagger 2.0 implementation for go. (Apache-2.0 license)
- [golang-migrate - migrate](https://github.com/golang-migrate/migrate) Database migrations. CLI and Golang library. (MIT license)
- [sqlc-dev - sqlc](https://github.com/sqlc-dev/sqlc) Generate type-safe code from SQL. (MIT license)
- [anton putra - nginx exporter](https://github.com/antonputra/tutorials/tree/main/lessons/141/prometheus-nginx-exporter) Best Place for DevOps. (MIT license)
