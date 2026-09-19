# BrewFlow

Coffee shop POS and inventory management system.

BrewFlow is a backend-first MVP focused on reliable order processing, inventory tracking, payments, and staff management. The project is being built as a solo-developer project with a focus on simple architecture, clear business rules, and a path toward a production-ready backend.

## Status

**Current phase:** MVP — In Progress

The current priority is completing the core order workflow and connecting it with inventory, payments, and discounts.

The project is intentionally being built in stages:

```text
Core MVP
   ↓
Production-ready MVP
   ↓
Future product features
```

---

## Tech Stack

### Backend

- Go 1.27+
- Gin — HTTP framework
- Bun — ORM / SQL toolkit
- PostgreSQL 17 — database
- golang-migrate — database migrations
- JWT (HS256) — authentication
- Air — development live reload
- Task — development commands

### Planned Frontend

- Next.js
- TypeScript
- Zod
- Drizzle

The frontend is intentionally planned after the backend API and core business workflows are established.

---

## Architecture

BrewFlow uses a simple layered architecture inspired by Clean Architecture principles:

```text
HTTP Handler
     ↓
Service / Business Logic
     ↓
Repository
     ↓
PostgreSQL
```

### Responsibilities

**Handlers**

- Parse HTTP requests
- Validate request-level input
- Extract authentication information
- Call services
- Return HTTP responses

Handlers do not contain business logic.

**Services**

- Contain business rules
- Coordinate multiple repositories/services
- Manage use-case transactions
- Return application DTOs and domain errors

**Repositories**

- Handle database queries
- Persist and retrieve data
- Do not contain business workflows

**Application / DI**

Dependencies are wired manually from the application composition root in:

```text
internal/app/container.go
```

This keeps dependencies explicit without introducing a dependency injection framework.

---

## Current MVP Scope

| Domain              | Features                                                                         |
| ------------------- | -------------------------------------------------------------------------------- |
| Authentication      | JWT login/logout, token blacklist, role-based access                             |
| Users               | List users                                                                       |
| Staff Onboarding    | Invitations, accept invitation, set password, cancel, list                       |
| Categories          | Create, rename, activate/deactivate, list                                        |
| Products            | Create, automatic SKU generation, find by ID/SKU, list, price in cents, category |
| Inventory           | Stock level per product                                                          |
| Inventory Movements | Received, sold, returned, damaged, adjusted                                      |
| Orders              | Create, list, detail, item management, status lifecycle                          |
| Payments            | Cash, card, mobile payments, cash change calculation                             |
| Discounts           | Create, list, percentage/fixed discounts                                         |

### Roles

- `owner`
- `manager`
- `staff`

Role permissions are enforced through protected routes and middleware.

---

## Development Plan

### Phase 1 — Core MVP

**Current**

Focus on completing the core POS workflow:

- Orders
- Order items
- Order status
- Inventory integration
- Inventory movements
- Payments
- Discounts
- Idempotent order creation

The goal is a complete end-to-end workflow rather than adding more features prematurely.

### Phase 2 — Production-Ready MVP

After the core workflows are complete:

- Unit and integration test suite
- Critical business workflow coverage
- Request validation
- Consistent API error responses
- Database constraints and indexes
- Transaction and concurrency hardening
- Authentication/security hardening
- Structured logging
- Configuration validation
- API documentation
- Observability
- Deployment configuration
- Backup/recovery considerations

The goal is to make the MVP reliable enough to operate as a real application, not simply feature-complete.

### Phase 3 — Future Features

Features intentionally outside the initial MVP:

- Customers / CRM
- Tables and dine-in floor management
- Shifts and cash drawer
- Receipt generation
- Sales reports and analytics
- Multi-location support

These will be added after the core POS workflow is stable.

---

## API Documentation

The API is currently documented through the route overview in this README.

As the API stabilizes, detailed API documentation will be maintained separately, with request/response examples, authentication requirements, validation rules, and error responses.

Planned documentation:

```text
docs/
└── api.md
```

An OpenAPI/Swagger specification may be added once the core API surface is stable.

---

## API Overview

All protected endpoints require authentication.

### Public

```text
POST /api/login
POST /api/logout
POST /api/accept-invitation
POST /api/set-password
```

### Users

```text
GET /api/protected/users
```

### Invitations

```text
POST /api/protected/invitations
GET  /api/protected/invitations
POST /api/protected/invitations/:id/cancel
```

### Categories

```text
POST   /api/protected/categories
GET    /api/protected/categories
PATCH  /api/protected/categories/:id
PATCH  /api/protected/categories/:id/status
```

### Products

```text
POST /api/protected/products
GET  /api/protected/products
GET  /api/protected/products/:id
GET  /api/protected/products/sku/:sku
```

### Inventory

```text
GET  /api/protected/inventory/:productId
GET  /api/protected/inventory/:productId/movements
POST /api/protected/inventory/:productId/receive
POST /api/protected/inventory/:productId/return
POST /api/protected/inventory/:productId/damage
POST /api/protected/inventory/:productId/adjust
```

### Orders

```text
POST   /api/protected/orders
GET    /api/protected/orders
GET    /api/protected/orders/:id
POST   /api/protected/orders/:id/status
POST   /api/protected/orders/:id/items
PATCH  /api/protected/orders/:id/items/:itemId
DELETE /api/protected/orders/:id/items/:itemId
POST   /api/protected/orders/:id/pay
POST   /api/protected/orders/:id/discount
```

### Discounts

```text
POST /api/protected/discounts
GET  /api/protected/discounts
```

The complete API reference will be maintained separately as the API grows.

---

## Key Engineering Decisions

### Money

All monetary values are stored as `BIGINT` cents.

No floating-point values are used for money.

### Tax

Tax is calculated by the backend using the configured `TAX_RATE` environment variable.

The default rate is `0.08`.

Tax is rounded half-up to the nearest cent.

### Orders and Inventory

Inventory changes are performed atomically with the related business operation.

Inventory movements are used to record stock changes:

```text
RECEIVED
SOLD
RETURNED
DAMAGED
ADJUSTED
```

Order-related inventory changes are handled by the order workflow rather than exposing a separate `sell` inventory endpoint.

### Transactions

Use cases that modify multiple resources are executed inside a single database transaction.

For example:

```text
Create Order
    ↓
Create Order Items
    ↓
Update Inventory
    ↓
Create Inventory Movements
```

If any operation fails, the transaction is rolled back.

### Idempotency

Order creation supports an optional:

```text
Idempotency-Key
```

header to prevent accidental duplicate order creation when clients retry requests.

### Errors

Services use package-level sentinel errors and domain/application errors.

HTTP error responses are mapped centrally rather than implementing HTTP-specific error handling throughout the service layer.

---

## Testing

Automated testing is planned as part of the production-ready MVP phase.

The testing strategy will focus on the parts of the system where correctness matters most:

- Service/business logic
- Repository/database behavior
- HTTP handlers
- Authentication and authorization
- Transactional workflows
- Inventory changes
- Order creation and lifecycle
- Payment and discount calculations

The goal is to prioritize meaningful tests around business rules and critical workflows rather than maximizing test coverage for its own sake.

Tests will be run with:

```bash
go test ./...
```

---

## Project Structure

```text
internal/
├── app/                       # Application composition and DI
├── config/                    # Configuration
├── database/                  # Database connection and transactions
├── middleware/                # Authentication, authorization, errors, logging
├── server/                    # HTTP server and routes
└── services/
    ├── auth/                  # JWT and token blacklist
    ├── user/                  # User management
    ├── account/               # Account relationships
    ├── invitations/           # Staff onboarding
    ├── categories/            # Product categories
    ├── products/              # Product catalog
    ├── inventory/             # Stock levels
    ├── inventory_movements/   # Stock movement history
    ├── orders/                # Orders and order items
    ├── payments/              # Payments
    └── discounts/             # Discounts

shared/                        # Shared non-domain utilities
db/
└── migrations/                # Database migrations
```

Each domain keeps its business logic close to the domain:

```text
service.go
repository.go
handler.go
models / DTOs
errors
```

depending on the needs of that domain.

---

## Getting Started

### Requirements

- Go
- Docker
- Task
- PostgreSQL

### Start PostgreSQL

```bash
task docker-up
```

### Run migrations

```bash
task migrate-up
```

### Start the development server

```bash
task dev
```

The API will be available at:

```text
http://localhost:8080
```

Health check:

```text
GET /health
```

---

## Development Principles

BrewFlow intentionally favors simple solutions over unnecessary abstraction.

- Keep business logic in services.
- Keep HTTP concerns in handlers.
- Keep database access in repositories.
- Return DTOs instead of database models from service boundaries.
- Use database transactions for multi-step state changes.
- Prefer explicit dependencies through manual DI.
- Keep domain rules close to the domain that owns them.
- Add abstractions when they solve a real problem, not preemptively.
- Keep migrations reversible.
- Prefer correctness and maintainability over premature optimization.

---

## License

MIT
