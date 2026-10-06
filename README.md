# BrewFlow

Backend-first coffee shop POS and inventory management system built with Go.

BrewFlow is a solo-developer project focused on learning and applying practical backend engineering: business rules, transactions, inventory consistency, authentication, payments, and clean separation between HTTP, application logic, and data access.

The project intentionally favors **simple, explicit architecture over unnecessary abstraction**.

> **Status:** MVP — In Progress

---

## What BrewFlow Does

BrewFlow is designed around the core workflow of a small coffee shop:

- Staff authentication and onboarding
- Product and category management
- Inventory tracking and stock adjustments
- Inventory movement history
- Order creation and order items
- Payment checkout through PayMongo
- Payment webhook processing
- Role-based access for protected staff operations
- Discounts and other POS rules as the MVP evolves

The current focus is completing the **order → inventory → payment** workflow before expanding into larger POS features.

---

## Tech Stack

### Backend

| Technology                                                  | Purpose                        |
| ----------------------------------------------------------- | ------------------------------ |
| [Go](https://go.dev/)                                       | Backend language               |
| [Gin](https://gin-gonic.com/)                               | HTTP framework                 |
| [Bun](https://bun.uptrace.dev/)                             | PostgreSQL ORM / SQL toolkit   |
| [PostgreSQL](https://www.postgresql.org/)                   | Primary database               |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Database migrations            |
| JWT                                                         | Authentication                 |
| [Air](https://github.com/air-verse/air)                     | Development live reload        |
| [Task](https://taskfile.dev/)                               | Development commands           |
| [Docker](https://www.docker.com/)                           | Local PostgreSQL environment   |
| `log/slog`                                                  | Structured application logging |

### Payment

- **PayMongo Hosted Checkout**
- Checkout creation through a payment gateway interface
- Webhook signature verification
- Webhook payload parsing
- Idempotent handling of already-paid payments

### Planned Frontend

The frontend is intentionally being developed after the backend workflows are stable.

- Next.js
- TypeScript
- Zod
- Drizzle

---

## Architecture

BrewFlow uses a straightforward layered architecture inspired by Clean Architecture.

```text
HTTP Request
     │
     ▼
 Handler
     │
     ▼
 Service / Business Logic
     │
     ├──────────────► Other Services
     │
     ▼
 Repository
     │
     ▼
 PostgreSQL
```

External services are accessed through small interfaces so business logic does not need to know provider-specific implementation details.

For example:

```text
Payment Service
      │
      ▼
PaymentGateway interface
      │
      ▼
PayMongo Adapter
      │
      ▼
PayMongo API
```

### Layer Responsibilities

**Handlers**

- Receive HTTP requests
- Parse and validate request input
- Extract request/authentication context
- Call application services
- Return HTTP responses

Handlers should not contain business workflows.

**Services**

- Own business rules and use cases
- Coordinate multiple services/repositories
- Validate business conditions
- Control transaction boundaries where needed
- Return application-level results and errors

**Repositories**

- Read and write database data
- Receive `bun.IDB` so they can work with either the normal database connection or a transaction
- Keep SQL/database details out of business logic

**Middleware**

- Request IDs
- Request logging
- Authentication
- Authorization
- Centralized error handling
- Panic recovery

**Application Composition**

Dependencies are wired manually in the application layer rather than using a dependency injection framework.

This keeps the dependency graph visible and easy to follow for a small project.

---

## Project Structure

```text
.
├── cmd/
│   └── api/                       # Application entry point
│
├── db/
│   └── migrations/                # PostgreSQL migrations
│
├── internal/
│   ├── app/                       # Application composition / dependency wiring
│   ├── config/                    # Environment configuration
│   ├── database/                  # PostgreSQL connection and transactions
│   ├── middleware/                # HTTP middleware
│   ├── server/                    # HTTP server and route registration
│   │
│   └── services/
│       ├── account/               # Account relationships
│       ├── auth/                  # JWT and token blacklist
│       ├── categories/            # Product categories
│       ├── inventory/             # Stock levels and stock operations
│       ├── inventory_movements/   # Inventory audit history
│       ├── invitations/           # Staff onboarding
│       ├── order_items/           # Order item operations
│       ├── orders/                # Order business logic
│       ├── payments/              # Payment workflow and gateway integration
│       ├── products/              # Product catalog
│       └── user/                  # User management
│
└── shared/
    └── logger/                    # Shared logging utilities
```

Most domains follow the same basic structure when it makes sense:

```text
model.go
repository.go
service.go
handler.go
errors.go
```

The project does not force every domain into identical abstractions when they are not needed.

---

## Current MVP

### Authentication & Staff

- JWT authentication
- Logout through token blacklist
- Staff invitations
- Invitation acceptance
- Password setup
- Protected API routes
- Owner / manager / staff roles

### Products & Categories

- Create products
- List products
- Find product by ID
- Find product by SKU
- Product categories
- Product activation/status management
- Automatic SKU generation
- Money stored as integer cents

### Inventory

- Inventory record per product
- Receive stock
- Damage stock
- Adjust stock
- Stock changes performed as part of order workflows
- Inventory movement records

Supported movement types include:

```text
RECEIVED
SOLD
RETURNED
DAMAGED
ADJUSTED
```

### Orders

- Create orders
- Order items
- Product price snapshots at order creation
- Inventory integration
- Order totals
- Pending → paid workflow
- Order numbers

### Payments

BrewFlow currently uses **PayMongo Hosted Checkout**.

The basic flow is:

```text
Client
  │
  │ Create checkout
  ▼
BrewFlow API
  │
  │ CreateCheckout()
  ▼
Payment Gateway Interface
  │
  ▼
PayMongo
  │
  │ Checkout URL
  ▼
Client
  │
  │ Customer completes payment
  ▼
PayMongo
  │
  │ Webhook
  ▼
BrewFlow
  │
  ├── Verify signature
  ├── Parse webhook
  ├── Find payment
  ├── Mark payment as PAID
  └── Mark order as PAID
```

The application stores its own payment record rather than relying only on PayMongo's data.

Current payment statuses:

```text
PENDING
PAID
FAILED
CANCELLED
REFUNDED
```

Current supported PayMongo payment methods represented by the domain model:

```text
CARD
GCASH
GRABPAY
MAYA
QRPH
```

### Discounts

The discount domain is part of the MVP and is being integrated into the order workflow.

---

## Important Engineering Decisions

### Money is stored as integer cents

Money is represented using `int64` rather than floating-point numbers.

For example:

```text
₱150.50 → 15050
```

This avoids floating-point rounding problems in financial calculations.

### Transactions

Multi-step operations use database transactions when the changes must succeed or fail together.

For example, order creation coordinates:

```text
Validate products
      ↓
Validate stock
      ↓
Begin transaction
      ↓
Update inventory
      ↓
Create order
      ↓
Create order items
      ↓
Commit
```

If a step fails, the transaction is rolled back.

The project uses a small transaction manager abstraction rather than manually managing transaction lifecycle in every service.

### Repository interfaces

Services depend on repository interfaces rather than concrete repository implementations.

This keeps database concerns separate from business logic and makes the important dependencies explicit.

### Payment gateway abstraction

The payment service does not directly depend on PayMongo's HTTP API.

Instead:

```text
Payment Service
      ↓
PaymentGateway
      ↓
PayMongo Adapter
```

This makes the payment provider replaceable without moving provider-specific code into the payment business logic.

### Webhook verification

PayMongo webhooks are not trusted simply because they reach the webhook endpoint.

BrewFlow:

1. Reads the raw request body.
2. Reads the PayMongo signature header.
3. Verifies the signature.
4. Parses the verified payload.
5. Finds the internal payment using the provider checkout ID.
6. Updates the payment.
7. Marks the related order as paid.

Already-paid payments are ignored so repeated successful webhook deliveries do not repeat the payment transition.

### Explicit dependency wiring

Dependencies are assembled manually in the application container.

There is no DI framework.

The goal is to keep the project easy to understand while still respecting dependency inversion where it provides a real benefit.

---

## API

The API is currently versionless and uses the `/api` prefix.

### Health

```http
GET /health
```

### Public Authentication

```http
POST /api/login
POST /api/logout
```

### Public Staff Onboarding

```http
POST /api/accept-invitation
POST /api/set-password
```

### PayMongo Webhook

This endpoint is public because PayMongo needs to call it directly.

```http
POST /api/webhooks/paymongo
```

The webhook is protected by PayMongo signature verification rather than JWT authentication.

### Protected Users

```http
GET /api/protected/users
```

### Protected Invitations

```http
POST /api/protected/invitations
GET  /api/protected/invitations
POST /api/protected/invitations/:id/cancel
```

### Protected Categories

```http
POST  /api/protected/categories
PATCH /api/protected/categories/:id
PATCH /api/protected/categories/:id/status
GET   /api/protected/categories
```

### Protected Products

```http
POST /api/protected/products
GET  /api/protected/products
GET  /api/protected/products/:id
GET  /api/protected/products/sku/:sku
```

### Protected Inventory

```http
GET  /api/protected/inventory/:productId
POST /api/protected/inventory/:productId/adjust
POST /api/protected/inventory/:productId/receive
POST /api/protected/inventory/:productId/damage
```

### Protected Orders

```http
POST /api/protected/orders
```

Additional order operations are being added as the order workflow is completed.

### Protected Payments

```http
POST /api/protected/payments/checkout
```

The checkout endpoint creates a PayMongo hosted checkout session for a pending order and returns the checkout URL to the client.

---

## Local Development

### Requirements

- Go 1.26.5+
- Docker
- Task
- PostgreSQL 17 (Docker is recommended for local development)

### 1. Clone the repository

```bash
git clone https://github.com/peter-bondad/gobrewflow.git
cd gobrewflow
```

### 2. Configure environment variables

Copy the example environment file:

```bash
cp .env.example .env
```

Set the required values for your local environment.

For local PostgreSQL, the Docker Compose setup uses port `5433` by default.

### 3. Install development tools

```bash
task setup
```

This installs the migration CLI and Air.

### 4. Start PostgreSQL

```bash
task docker-up
```

### 5. Run migrations

```bash
task migrate-up
```

### 6. Start the API

```bash
task dev
```

The server uses the `PORT` value from your environment.

### Useful commands

```bash
# Start development server
task dev

# Start PostgreSQL
task docker-up

# Stop PostgreSQL and remove its volume
task docker-down

# Run pending migrations
task migrate-up

# Roll back the latest migration
task migrate-down

# Reset the local database and run migrations again
task migrate-refresh

# Create a migration
task migrate-create NAME=create_example_table

# Run tests
go test ./...
```

---

## Environment Variables

The application reads configuration from environment variables.

The available configuration includes:

```text
APP_ENV
PORT
LOG_LEVEL
APP_SHUTDOWN_TIMEOUT

DB_HOST
DB_PORT
DB_USER
DB_PASSWORD
DB_NAME
DB_SSL_MODE

JWT_SECRET_KEY
INVITATION_BASE_URL
INVITATION_TTL

PAYMONGO_BASE_URL
PAYMONGO_SUCCESS_URL
PAYMONGO_CANCEL_URL
PAYMONGO_TEST_PUBLIC_KEY
PAYMONGO_TEST_SECRET_KEY
```

See `.env.example` for the current configuration surface.

**Never commit real credentials or API keys.**

---

## Testing

Testing is part of the production-readiness phase.

The priority is meaningful coverage of business-critical behavior rather than chasing a percentage:

- Order creation
- Inventory changes
- Transactional workflows
- Payment state transitions
- Webhook verification and parsing
- Authentication and authorization
- Repository behavior
- Validation and error handling

Run the current test suite with:

```bash
go test ./...
```

---

## Development Roadmap

### Phase 1 — Core MVP

**Current focus**

- [x] Authentication foundation
- [x] Staff invitations
- [x] Categories
- [x] Products
- [x] Inventory
- [x] Inventory movements
- [x] Basic order creation
- [x] PayMongo hosted checkout foundation
- [x] PayMongo webhook processing
- [ ] Complete order lifecycle
- [ ] Complete discount integration
- [ ] Finalize payment/order state transitions

### Phase 2 — Production Readiness

After the core workflow is stable:

- [ ] Unit tests
- [ ] Integration tests
- [ ] Critical workflow coverage
- [ ] Stronger request validation
- [ ] Consistent API error responses
- [ ] Database indexes and constraints review
- [ ] Transaction/concurrency hardening
- [ ] Authentication/security hardening
- [ ] Structured logging improvements
- [ ] API documentation
- [ ] Observability
- [ ] Deployment configuration
- [ ] Backup/recovery considerations

### Phase 3 — Future POS Features

Features intentionally kept outside the initial MVP:

- [ ] Customers / CRM
- [ ] Dine-in tables and floor management
- [ ] Staff shifts and cash drawer management
- [ ] Receipt generation
- [ ] Sales reports and analytics
- [ ] Multi-location support

The goal is to finish the core workflow before expanding the product surface.

---

## Development Principles

BrewFlow is a learning project, but the codebase is being built with production-oriented habits.

- **Keep it simple.**
- Put business rules in services.
- Keep HTTP concerns in handlers.
- Keep database access in repositories.
- Use interfaces at meaningful boundaries.
- Prefer explicit dependencies over framework magic.
- Use transactions for atomic multi-step operations.
- Keep provider-specific code behind adapters.
- Store money safely as integer cents.
- Prefer type-safe, explicit models.
- Avoid abstractions until they solve a real problem.
- Optimize for correctness and maintainability before performance.
- Build the backend workflow first, then add the frontend.

---

## Why This Project?

BrewFlow is primarily a backend engineering project.

It is being used to practice real backend concerns that are easy to overlook in simple CRUD applications:

- Designing business workflows
- Maintaining database consistency
- Handling transactions
- Managing inventory state
- Modeling payment states
- Integrating third-party payment providers
- Verifying webhooks
- Handling retries and duplicate events
- Separating application logic from infrastructure
- Designing APIs that can evolve without unnecessary complexity

The goal is not to build the largest POS system possible. The goal is to build a **small, understandable system correctly**.

---

## License

MIT License

See [LICENSE](LICENSE) for details.
