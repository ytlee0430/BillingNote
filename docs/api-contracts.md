# API Contracts

**Generated:** 2026-03-06
**Base URL:** `http://localhost:8080`

---

## Authentication

All protected endpoints require:
```
Authorization: Bearer <JWT_TOKEN>
```

JWT payload contains: `user_id`, `email`, `exp`

---

## Public Endpoints

### POST /api/auth/register

Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "User Name"
}
```

**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name"
  }
}
```

### POST /api/auth/login

Authenticate user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response (200):** Same as register response.

---

## Protected Endpoints

### GET /api/auth/me

Get current user info.

**Response (200):**
```json
{
  "user_id": 1,
  "email": "user@example.com"
}
```

---

### Transactions

#### GET /api/transactions

List transactions with filtering and pagination.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | - | Filter by `income` or `expense` |
| `start_date` | string | - | Start date (YYYY-MM-DD) |
| `end_date` | string | - | End date (YYYY-MM-DD) |
| `category_id` | int | - | Filter by category |
| `page` | int | 1 | Page number |
| `page_size` | int | 10 | Items per page |

**Response (200):**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 4,
      "amount": 150.00,
      "type": "expense",
      "description": "Lunch",
      "transaction_date": "2026-01-05T00:00:00Z",
      "source": "manual",
      "created_at": "2026-01-05T10:00:00Z",
      "updated_at": "2026-01-05T10:00:00Z",
      "category": {
        "id": 4,
        "name": "Dining",
        "type": "expense",
        "icon": "icon",
        "color": "#FF6B6B"
      }
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 10
}
```

#### POST /api/transactions

Create a new transaction.

**Request:**
```json
{
  "category_id": 4,
  "amount": 150.00,
  "type": "expense",
  "description": "Lunch",
  "transaction_date": "2026-01-05",
  "source": "manual"
}
```

**Response (201):** Transaction object.

#### GET /api/transactions/:id

Get a single transaction by ID.

#### PUT /api/transactions/:id

Update a transaction.

**Request:** Partial transaction fields (all optional).

#### DELETE /api/transactions/:id

Delete a transaction.

**Response (200):**
```json
{ "message": "transaction deleted successfully" }
```

---

### Statistics

#### GET /api/stats/monthly

Get monthly income/expense summary.

**Query Parameters:**
| Param | Type | Default |
|-------|------|---------|
| `year` | int | current year |
| `month` | int | current month |

#### GET /api/stats/category

Get category-wise statistics.

**Query Parameters:**
| Param | Type | Required |
|-------|------|----------|
| `start_date` | string (YYYY-MM-DD) | Yes |
| `end_date` | string (YYYY-MM-DD) | Yes |
| `type` | string | No (income/expense) |

---

### Categories

#### GET /api/categories

Get all categories.

#### GET /api/categories/type/:type

Get categories by type (`income` or `expense`).

---

### PDF Upload

#### POST /api/upload/pdf

Upload and parse PDF credit card statements.

**Content-Type:** `multipart/form-data`
**Field:** `files` (multiple PDF files)

**Response (200):**
```json
{
  "results": [
    {
      "filename": "statement.pdf",
      "bank": "Cathay",
      "transactions": [...],
      "total_amount": 15000.00,
      "error": ""
    }
  ]
}
```

#### POST /api/transactions/import

Import parsed transactions from PDF preview.

**Request:**
```json
{
  "transactions": [
    {
      "date": "2026-01-05",
      "description": "Store purchase",
      "amount": 500.00,
      "category": "Shopping"
    }
  ]
}
```

---

### PDF Password Settings

#### GET /api/settings/pdf-passwords

List user's PDF passwords (without actual password values).

#### POST /api/settings/pdf-passwords

Set a single PDF password.

**Request:**
```json
{
  "password": "secret123",
  "priority": 1,
  "label": "ID last 4 digits"
}
```

#### PUT /api/settings/pdf-passwords

Set multiple PDF passwords at once.

#### DELETE /api/settings/pdf-passwords/:priority

Delete a PDF password by priority slot (1-4).

---

## Error Response Format

All errors follow a consistent format:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {},
  "trace_id": "request-uuid"
}
```

Common HTTP status codes:
- `400` - Validation error
- `401` - Unauthorized (missing/invalid JWT)
- `403` - Forbidden
- `404` - Resource not found
- `500` - Internal server error

---

## Health Check

### GET /health

**Response (200):**
```json
{ "status": "ok" }
```
