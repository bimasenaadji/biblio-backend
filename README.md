# 📚 Biblio - Smart Physical Library & Book Tracking Backend

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Framework](https://img.shields.io/badge/Fiber-v2-00ACD7?style=flat&logo=go)](https://gofiber.io)
[![ORM](https://img.shields.io/badge/GORM-PostgreSQL-blue?style=flat&logo=postgresql)](https://gorm.io)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal%20Ports%20%26%20Adapters-orange)](#-hexagonal-architecture)
[![API Documentation](https://img.shields.io/badge/Swagger-OpenAPI%202.0-85EA2D?style=flat&logo=swagger)](http://localhost:3000/swagger/index.html)

**Biblio** is an enterprise-grade backend service built for physical library operations, barcode-driven book circulation, and inventory audits. Designed with a strict **4-Layer Hexagonal Architecture (Ports & Adapters)**, it ensures high testability, domain decoupling, clean concurrency handling, and robust data persistence.

---

## 🏗️ Hexagonal Architecture

Biblio follows a strict separation of concerns where the core business rules never depend on external frameworks:

```text
               ┌────────────────────────────────────────────────────────┐
               │                  Primary Adapters                      │
               │  [Fiber HTTP Handler] ──> [API DTO Coordinator]        │
               └───────────────────────────┬────────────────────────────┘
                                           │ (calls Inbound Port)
                                           ▼
               ┌────────────────────────────────────────────────────────┐
               │                     Core Domain                        │
               │  [Service Port (Interface)]                            │
               │              │                                         │
               │              ▼                                         │
               │  [Usecase Service] ─── operates on ───> [Entities]     │
               │              │                                         │
               │              ▼                                         │
               │  [Repository Port (Interface)]                         │
               └───────────────────────────┬────────────────────────────┘
                                           │ (implemented by Outbound Port)
                                           ▼
               ┌────────────────────────────────────────────────────────┐
               │                 Secondary Adapters                     │
               │  [GORM PostgreSQL Repository]                          │
               └────────────────────────────────────────────────────────┘
```

### Layer Breakdown

1. **`core/entities/`**: Core domain data structures (`Book`) with GORM hooks and zero HTTP awareness.
2. **`ports/`**: Go interfaces specifying business use cases (`BookService`) and persistence contracts (`BookRepository`).
3. **`core/usecases/`**: Intelligent orchestrator enforcing business rules (e.g., verify existence prior to updates/deletions).
4. **`adapters/secondary/postgres/`**: Data layer implementing `ports.BookRepository` via GORM.
5. **`adapters/primary/http/`**:
   - `dto.go`: Request validation and response mapping models.
   - `api.go`: Bidirectional mapper between DTOs and Domain Entities.
   - `handler.go`: Fiber HTTP request binding, status codes, and Swagger annotations.

---

## 🛠️ Tech Stack

- **Language**: Go 1.22+
- **HTTP Framework**: [Go Fiber v2](https://github.com/gofiber/fiber)
- **Database & ORM**: PostgreSQL 16 (isolated on port `5433`) & [GORM](https://gorm.io)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **API Documentation**: [Swaggo / Swagger](https://github.com/swaggo/swag)
- **Infrastructure**: Docker Compose (`PostgreSQL`, `MinIO S3`, `Jaeger`, `Grafana`)

---

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed on your machine:
- [Go (1.22 or newer)](https://golang.org/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Swag CLI](https://github.com/swaggo/swag):
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

---

### Step 1: Environment Setup

Clone the repository and create your local `.env` file from `.env.example`:

```bash
cp .env.example .env
```

Verify that `DB_PORT=5433` is set to connect to the dedicated Biblio PostgreSQL container without colliding with host port 5432.

---

### Step 2: Start Infrastructure Services

Spin up the standalone container stack (PostgreSQL, MinIO, Jaeger, Grafana):

```bash
make docker-up
# or manually: docker compose up -d
```

---

### Step 3: Run the Application

Generate fresh Swagger documentation and start the server:

```bash
make dev
# or manually:
# swag init -g cmd/api/main.go -o docs
# go run cmd/api/main.go
```

The server will automatically run database schema migrations and listen on `http://localhost:3000`.

---

## 📖 API Documentation & Swagger UI

Once the application is running, visit the interactive Swagger UI:

🔗 **[http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)**

### Available Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Server health check and version metadata |
| `GET` | `/swagger/*` | Interactive Swagger API documentation |
| `POST` | `/api/v1/books` | Create a new book catalog title |
| `GET` | `/api/v1/books` | Retrieve list of all book titles |
| `GET` | `/api/v1/books/:id` | Retrieve single book title by UUID |
| `PUT` | `/api/v1/books/:id` | Update an existing book title by UUID |
| `DELETE` | `/api/v1/books/:id` | Soft-delete a book record by UUID |

---

## 📦 Project Structure

```text
biblio/
├── .agents/                    # Architecture rules & PRD specifications
├── cmd/
│   └── api/
│       └── main.go             # Application entrypoint, DI wiring & graceful shutdown
├── docs/                       # Auto-generated OpenAPI / Swagger specs
├── internal/
│   ├── book/
│   │   ├── core/
│   │   │   ├── entities/       # Domain models (Book)
│   │   │   └── usecases/       # Business use cases (BookService)
│   │   ├── ports/              # Inbound & Outbound interfaces
│   │   └── adapters/
│   │       ├── primary/http/   # HTTP handlers, DTOs & API coordinator
│   │       └── secondary/postgres/ # PostgreSQL GORM repository
│   ├── infrastructure/
│   │   ├── config/             # Viper configuration loader
│   │   └── database.go         # GORM connection pool & SQL logger
│   └── utils/
│       └── response/           # Unified APIResponse envelope
├── .env.example                # Sample environment configuration
├── .gitignore                  # Git ignore rules
├── docker-compose.yaml         # Standalone PostgreSQL, MinIO, Jaeger stack
├── go.mod
├── go.sum
├── makefile                    # Automation commands (dev, build, swagger)
└── README.md                   # Project documentation
```

---

## 🛡️ Graceful Shutdown

Biblio handles OS interrupt signals (`SIGINT`, `SIGTERM`) gracefully:
- Stops accepting new incoming HTTP connections immediately.
- Grants up to **10 seconds** for pending requests to finish processing.
- Cleanly closes the PostgreSQL connection pool without connection leaks or data corruption.

Trigger it anytime during local development by pressing `Ctrl + C` in the server terminal.

---

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.
