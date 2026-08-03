# Todo API — Go + PostgreSQL + Redis

REST API untuk manajemen tugas (To-Do List) dibangun dengan Go, PostgreSQL, dan Redis.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21+ |
| HTTP Framework | Gin |
| Database | PostgreSQL 16 |
| ORM | GORM |
| Caching | Redis 7 |
| Authentication | JWT (golang-jwt/jwt v5) |
| Logging | zerolog |
| Testing | Go testing + testify |

## Fitur

- ✅ CRUD Tasks (Create, Read, Update, Delete)
- ✅ Filter, Pagination, dan Search
- ✅ Input Validation
- ✅ Error Handling terstruktur
- ✅ Concurrent Execution (goroutines + WaitGroup)
- ✅ Redis Caching (dengan fallback jika Redis tidak tersedia)
- ✅ JWT Authentication
- ✅ Structured Logging (zerolog)
- ✅ Unit Tests (testify + mock)
- ✅ Graceful Shutdown
- ✅ Docker Compose untuk PostgreSQL & Redis

## Persyaratan

- Go 1.21+
- Docker & Docker Compose
- Git

## Cara Menjalankan

### 1. Clone & Masuk ke Direktori

```bash
git clone <repository-url>
cd backend
```

### 2. Salin File Konfigurasi

```bash
cp .env.example .env
```

Edit `.env` sesuai konfigurasi lokal kamu:

```env
# Application
APP_PORT=8082
APP_ENV=development

# PostgreSQL
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=<your_db_password>
DB_NAME=todoapp
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<your_redis_password_or_blank>
REDIS_DB=0

# JWT
JWT_SECRET=<your_jwt_secret_key>
JWT_EXPIRATION_HOURS=24
```

### 3. Jalankan PostgreSQL & Redis via Docker

```bash
docker compose up -d
```

> Ini akan otomatis menjalankan PostgreSQL di port 5433 dan Redis di port 6379.
> Migration tabel akan dijalankan otomatis saat container pertama kali dibuat.

### 4. Jalankan Server

```bash
go run ./cmd/server/main.go
```

Server berjalan di: `http://localhost:8082`

### 5. Cek Health

```bash
curl http://localhost:8082/health
```

---

## API Documentation (Swagger)

Proyek ini menyediakan dokumentasi interaktif untuk seluruh API endpoint menggunakan Swagger UI.
Setelah server berhasil dijalankan, kamu dapat mengakses dokumentasi tersebut melalui browser pada alamat berikut:

- **Swagger UI:** [http://localhost:8082/docs/index.html](http://localhost:8082/docs/index.html)

---

## API Endpoints

### Auth (Publik)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/auth/register` | Daftar pengguna baru |
| POST | `/api/v1/auth/login` | Login dan dapatkan JWT token |

### Tasks (Membutuhkan JWT)

Tambahkan header `Authorization: Bearer <token>` untuk setiap request task.

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/tasks` | Buat task baru |
| GET | `/api/v1/tasks` | Ambil semua tasks (filter, paginasi, search) |
| GET | `/api/v1/tasks/:id` | Ambil task berdasarkan ID |
| PUT | `/api/v1/tasks/:id` | Update task |
| DELETE | `/api/v1/tasks/:id` | Hapus task |

---

## Contoh Request & Response

### Register

```bash
curl -X POST http://localhost:8082/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8082/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'
```

Response:
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-08-04T22:00:00Z",
    "user": { "id": "...", "name": "John Doe", "email": "john@example.com" }
  }
}
```

### Create Task

```bash
curl -X POST http://localhost:8082/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Belajar Go","description":"Pelajari goroutines","status":"pending","due_date":"2026-08-31"}'
```

### Get All Tasks (dengan filter & paginasi)

```bash
curl "http://localhost:8082/api/v1/tasks?status=pending&page=1&limit=10&search=belajar" \
  -H "Authorization: Bearer <token>"
```

Response:
```json
{
  "status": "success",
  "data": {
    "tasks": [
      {
        "id": "...",
        "title": "Belajar Go",
        "description": "Pelajari goroutines",
        "status": "pending",
        "due_date": "2026-08-31",
        "created_at": "...",
        "updated_at": "..."
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_tasks": 1,
      "limit": 10
    }
  }
}
```

### Get Task by ID

```bash
curl http://localhost:8082/api/v1/tasks/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <token>"
```

### Update Task

```bash
curl -X PUT http://localhost:8082/api/v1/tasks/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Belajar Go Lanjut","description":"Pelajari channels & concurrency","status":"completed","due_date":"2026-09-05"}'
```

Response:
```json
{
  "status": "success",
  "message": "Task updated successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Belajar Go Lanjut",
    "description": "Pelajari channels & concurrency",
    "status": "completed",
    "due_date": "2026-09-05",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### Delete Task

```bash
curl -X DELETE http://localhost:8082/api/v1/tasks/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <token>"
```

Response:
```json
{
  "status": "success",
  "message": "Task deleted successfully"
}
```

---

## Menjalankan Unit Tests

```bash
go test ./tests/... -v
```

Atau untuk melihat coverage:

```bash
go test ./tests/... -v -cover
```

---

## Struktur Proyek

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Konfigurasi aplikasi
│   ├── database/
│   │   ├── postgres.go          # Koneksi PostgreSQL
│   │   └── redis.go             # Koneksi Redis
│   ├── middleware/
│   │   ├── auth.go              # JWT middleware
│   │   └── logger.go            # Request logger
│   ├── models/
│   │   ├── task.go              # Task model
│   │   └── user.go              # User model
│   ├── handlers/
│   │   ├── task_handler.go      # Task HTTP handlers
│   │   ├── auth_handler.go      # Auth HTTP handlers
│   │   └── helpers.go           # Shared response helpers
│   ├── repository/
│   │   ├── task_repository.go           # Interface
│   │   ├── task_repository_impl.go      # Implementasi GORM
│   │   └── user_repository.go           # User repository
│   ├── service/
│   │   ├── task_service.go      # Business logic + concurrency
│   │   └── auth_service.go      # Auth + JWT logic
│   └── dto/
│       └── task_dto.go          # Request/Response DTOs
├── migrations/
│   └── 001_create_tables.sql    # SQL migration
├── tests/
│   └── task_service_test.go     # Unit tests
├── .env                         # Konfigurasi lokal (tidak di-commit)
├── .env.example                 # Template konfigurasi
├── docker-compose.yml           # PostgreSQL + Redis
├── go.mod
└── README.md
```

## Concurrent Execution

Concurrent execution diterapkan di beberapa tempat:

1. **Server Startup** (`main.go`): Server HTTP berjalan di goroutine terpisah
2. **Cache Invalidation** (`task_service.go`): Setelah write operation (create/update/delete), cache diinvalidasi secara concurrent menggunakan `go func()` + `sync.WaitGroup`
3. **Async Cache Set** (`task_service.go`): Hasil query di-cache secara async tanpa memblokir response ke client

## Graceful Shutdown

Server mendengarkan sinyal `SIGINT` / `SIGTERM` dan memberi waktu 10 detik untuk menyelesaikan request yang sedang berjalan sebelum shutdown.
