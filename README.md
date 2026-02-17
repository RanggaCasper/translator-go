# Subtitle Translator API

A high-performance subtitle translation service built with Go and Fiber framework. This API translates VTT and ASS subtitle files to Indonesian (or any target language) using Google Translate, with **MySQL database for metadata** and **file system for content storage**.

Current release: [`v1.0.0`](./changelog.md)

## Features

- 🚀 Fast and efficient subtitle translation
- 📝 Supports VTT and ASS subtitle formats
- 🗄️ **Hybrid storage**: MySQL for metadata + File system for content
- 💾 Subtitle content saved as `.vtt` files
- 📊 Database stores metadata (URL, language, file path, timestamps)
- 🔄 Full CRUD operations (Create, Read, Update, Delete)
- 🌐 Batch translation for better performance
- 🎯 Informal Indonesian translation style (anime-friendly)
- 🔧 Modular and clean architecture
- 📄 Pagination support

## Storage Architecture

### Hybrid Approach: Best of Both Worlds

```
MySQL Database (Metadata)          File System (Content)
┌─────────────────────────┐        ┌──────────────────────┐
│ id                      │        │ storage/subtitles/   │
│ subtitle_id             │───────>│  ├── abc123.vtt      │
│ url                     │        │  ├── def456.vtt      │
│ target_lang             │        │  └── ghi789.vtt      │
│ file_path ──────────────┼────┐   └──────────────────────┘
│ file_size               │    │
│ created_at              │    └──> Points to actual file
│ updated_at              │
└─────────────────────────┘
```

### Why Hybrid Storage?

✅ **Fast Queries** - Database indexes for quick searches by language, URL, date
✅ **Efficient Storage** - Large subtitle content in files, not database
✅ **Easy Backup** - Database for structure, files for content
✅ **Best Performance** - Database for metadata queries, direct file access for content
✅ **Scalability** - Can move files to CDN/S3 without changing database
✅ **Simple CRUD** - Easy to list, filter, sort via database queries

## Architecture

```
subtitle-translator-go/
├── config/              # Database configuration
├── internal/
│   ├── handler/        # HTTP handlers
│   ├── models/         # Database models (GORM)
│   ├── repository/     # Data access (DB + File operations)
│   ├── routes/         # Route definitions
│   └── service/        # Business logic
├── pkg/
│   ├── translator/     # Translation logic
│   └── utils/          # Utilities
├── storage/            # Subtitle files (auto-created)
│   └── subtitles/      # VTT content files
├── main.go
└── README.md
```

## Installation

### Prerequisites

- Go 1.21 or higher
- MySQL 5.7 or higher

### Setup

1. **Clone repository:**
```bash
git clone <repository-url>
cd subtitle-translator-go
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Create database:**
```sql
CREATE DATABASE subtitle_translator CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. **Configure environment:**
```bash
cp .env.example .env
nano .env  # Edit with your database credentials
```

5. **Run application:**
```bash
go run main.go
```

The server will start on `http://localhost:3000`

## API Endpoints

### Base URL
```
http://localhost:3000/api/v1
```

### Endpoints Table

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/subtitles/translate` | Translate subtitle (save to DB + file) |
| `POST` | `/subtitles/translate/text` | Translate a single sentence/text |
| `POST` | `/subtitles/translate/batch` | Translate many text blocks in one payload |
| `GET` | `/subtitles` | Get all subtitles (metadata only) |
| `GET` | `/subtitles/:id` | Get subtitle by ID (with content) |
| `PUT` | `/subtitles/:id` | Update subtitle file content |
| `DELETE` | `/subtitles/:id` | Delete subtitle (DB record + file) |

---

### 1. Health Check

Check if API is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": true,
  "data": {
    "message": "Subtitle Translator API is running",
    "version": "2.0.0"
  }
}
```

---

### 2. Translate Subtitle

Translate subtitle from URL. Saves metadata to database and content to file.

**Endpoint:** `POST /subtitles/translate`

**Request Body:**
```json
{
  "url": "https://example.com/subtitle.vtt",
  "format": "vtt",
  "target_lang": "id",
  "source_lang": "auto",
  "referer": "https://example.com"
}
```

**Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | Yes | - | URL of subtitle file |
| `format` | string | Yes | - | Format (`vtt` or `ass`) |
| `target_lang` | string | No | `id` | Target language code |
| `source_lang` | string | No | `auto` | Source language code |
| `referer` | string | No | - | HTTP Referer header |

**Success Response:**
```json
{
  "status": true,
  "data": {
    "id": 1,
    "subtitle_id": "a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4",
    "url": "https://example.com/subtitle.vtt",
    "target_lang": "id",
    "source_lang": "auto",
    "format": "vtt",
    "file_path": "storage/subtitles/a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4.vtt",
    "content": "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHalo, apa kabar?\n\n",
    "file_size": 512,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 2a. Translate Single Text

Translate one sentence/text quickly.

**Endpoint:** `POST /subtitles/translate/text`

**Request Body:**
```json
{
  "text": "Aku mau pergi ke pasar besok.",
  "target_lang": "en",
  "source_lang": "auto"
}
```

**Success Response:**
```json
{
  "status": true,
  "data": {
    "text": "Aku mau pergi ke pasar besok.",
    "translated_text": "I want to go to the market tomorrow.",
    "target_lang": "en",
    "source_lang": "auto"
  }
}
```

---

### 2b. Translate Batch Content

Translate many text blocks (`data.title` + `data.content[].text`) in one request.

**Endpoint:** `POST /subtitles/translate/batch`

**Request Body:**
```json
{
  "target_lang": "en",
  "source_lang": "auto",
  "data": {
    "title": "Tuan Misteri 2 Lingkaran Yang Tak Terhindarkan - Chapter 484",
    "content": [
      {
        "type": "heading",
        "text": "Chapter 484: Ejekan",
        "src": null
      },
      {
        "type": "text",
        "text": "Franca sama sekali tidak terkejut...",
        "src": null
      }
    ]
  }
}
```

**Notes:**
- Unknown fields are ignored (you can send larger wrapper JSON as long as it contains `data.title` or `data.content[].text`).
- Empty text values are skipped automatically.
- Built for large payloads (batch chunking + long-text fallback).

**Success Response:**
```json
{
  "status": true,
  "data": {
    "target_lang": "en",
    "source_lang": "auto",
    "translated_count": 3,
    "data": {
      "title": "Lord of Mysteries 2 Circle of Inevitability - Chapter 484",
      "content": [
        {
          "type": "heading",
          "text": "Chapter 484: Mockery",
          "src": null
        }
      ]
    }
  }
}
```

---

### 3. Get All Subtitles

Fetch all subtitle metadata with pagination. **Content not included** (faster queries).

**Endpoint:** `GET /subtitles`

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | `1` | Page number |
| `limit` | integer | `10` | Items per page (max: 100) |
| `target_lang` | string | - | Filter by language |

**Example:**
```bash
GET /api/v1/subtitles?page=1&limit=10&target_lang=id
```

**Success Response:**
```json
{
  "status": true,
  "data": [
    {
      "id": 1,
      "subtitle_id": "a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4",
      "url": "https://example.com/subtitle.vtt",
      "target_lang": "id",
      "source_lang": "auto",
      "format": "vtt",
      "file_path": "storage/subtitles/a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4.vtt",
      "file_size": 512,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "current_page": 1,
    "total_pages": 5,
    "next_page": 2,
    "has_next": true,
    "has_prev": false
  }
}
```

---

### 4. Get Subtitle by ID

Fetch single subtitle **with content loaded from file**.

**Endpoint:** `GET /subtitles/:id`

**Example:**
```bash
GET /api/v1/subtitles/1
```

**Success Response:**
```json
{
  "status": true,
  "data": {
    "id": 1,
    "subtitle_id": "a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4",
    "url": "https://example.com/subtitle.vtt",
    "target_lang": "id",
    "source_lang": "auto",
    "format": "vtt",
    "file_path": "storage/subtitles/a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4.vtt",
    "content": "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHalo, apa kabar?\n\n",
    "file_size": 512,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 5. Update Subtitle

Update subtitle file content and file size in database.

**Endpoint:** `PUT /subtitles/:id`

**Request Body:**
```json
{
  "content": "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHalo, gimana kabarnya?\n\n"
}
```

**Success Response:**
```json
{
  "status": true,
  "data": {
    "id": 1,
    "subtitle_id": "a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4",
    "url": "https://example.com/subtitle.vtt",
    "target_lang": "id",
    "source_lang": "auto",
    "format": "vtt",
    "file_path": "storage/subtitles/a7f5c1d2e3b4a5c6d7e8f9a0b1c2d3e4.vtt",
    "content": "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHalo, gimana kabarnya?\n\n",
    "file_size": 520,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  }
}
```

---

### 6. Delete Subtitle

Delete database record **and** file permanently.

**Endpoint:** `DELETE /subtitles/:id`

**Example:**
```bash
DELETE /api/v1/subtitles/1
```

**Success Response:**
```json
{
  "status": true,
  "data": {
    "message": "Subtitle deleted successfully"
  }
}
```

---

## Database Schema

### subtitles Table

```sql
CREATE TABLE `subtitles` (
  `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
  `subtitle_id` varchar(32) UNIQUE NOT NULL,
  `url` text NOT NULL,
  `target_lang` varchar(10) NOT NULL,
  `source_lang` varchar(10) NOT NULL,
  `format` varchar(10) NOT NULL,
  `file_path` varchar(500) NOT NULL,
  `file_size` bigint NOT NULL,
  `created_at` datetime(3),
  `updated_at` datetime(3),
  `deleted_at` datetime(3),
  INDEX idx_target_lang (target_lang),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Key Points:**
- `subtitle_id`: MD5 hash of URL + target_lang + format (for duplicate check)
- `file_path`: Path to actual VTT file in storage
- `file_size`: Size in bytes (for display/monitoring)

---

## Standard Response Format

### Success Response
```json
{
  "status": true,
  "data": { /* ... */ }
}
```

### Error Response
```json
{
  "status": false,
  "error": "Error type",
  "message": "Details"
}
```

### Paginated Response
```json
{
  "status": true,
  "data": [ /* ... */ ],
  "meta": {
    "current_page": 1,
    "total_pages": 10,
    "next_page": 2,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Translation Features

### Informal Indonesian Style

Automatically converts formal to informal:
- `Anda` → `kamu`, `Saya` → `Aku`
- `mengatakan` → `bilang`, `membuat` → `bikin`
- `tidak` → `nggak`, `terima kasih` → `makasih`

### Batch Translation

Processes up to 80 subtitle lines per request for optimal performance.

---

## cURL Examples

### Translate Subtitle
```bash
curl -X POST http://localhost:3000/api/v1/subtitles/translate \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/subtitle.vtt",
    "format": "vtt",
    "target_lang": "id"
  }'
```

### Get All Subtitles
```bash
curl http://localhost:3000/api/v1/subtitles?page=1&limit=10
```

### Get Subtitle by ID
```bash
curl http://localhost:3000/api/v1/subtitles/1
```

### Update Subtitle
```bash
curl -X PUT http://localhost:3000/api/v1/subtitles/1 \
  -H "Content-Type: application/json" \
  -d '{"content": "WEBVTT\n\n..."}'
```

### Delete Subtitle
```bash
curl -X DELETE http://localhost:3000/api/v1/subtitles/1
```

---

## Development

### Build
```bash
go build -o subtitle-translator main.go
```

### Run
```bash
./subtitle-translator
```

### Using Makefile
```bash
make build    # Build
make run      # Build and run
make clean    # Clean files
```

---

## Docker Deployment

### Using Docker Compose
```bash
docker-compose up -d
```

Includes MySQL container and volume persistence.

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3000` |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL username | `root` |
| `DB_PASSWORD` | MySQL password | - |
| `DB_NAME` | Database name | `subtitle_translator` |

---

## Performance

✅ **Fast metadata queries** - Database indexes
✅ **Efficient content storage** - Direct file system
✅ **No ORM overhead for files** - Pure Go file operations
✅ **Batch translation** - 80 lines per request
✅ **Connection pooling** - GORM manages connections
✅ **Permanent storage** - Translate once, use forever

---

## Backup Strategy

### Database Backup
```bash
mysqldump -u root -p subtitle_translator > backup.sql
```

### File Backup
```bash
tar -czf storage-backup.tar.gz storage/
```

### Restore
```bash
mysql -u root -p subtitle_translator < backup.sql
tar -xzf storage-backup.tar.gz
```

---

## Advantages of Hybrid Storage

| Aspect | Advantage |
|--------|-----------|
| **Query Speed** | Database indexes for fast searches |
| **Storage Efficiency** | Large content in files, not DB |
| **Scalability** | Can move files to CDN/S3 easily |
| **Performance** | Best of both worlds |
| **Maintenance** | Separate backups for data/content |
| **Flexibility** | Change storage backend without DB migration |

---

## Contributing

Contributions welcome! Please submit a Pull Request.

## License

MIT License

---

## Performance Optimization

### Speed Improvements

The Go version is optimized for maximum speed:

✅ **Concurrent Batch Processing** - Multiple chunks translated in parallel
✅ **HTTP Connection Pooling** - Reuses connections to Google Translate
✅ **Smart Retry Logic** - Auto-splits large batches if they fail
✅ **Concurrent Fallback** - Individual translations run in parallel
✅ **Fast Timeouts** - 10s timeout (vs 15s in original)
✅ **Progress Logging** - See translation progress in real-time

### Translation Flow

```
1. Parse subtitle → Extract text lines
2. Split into chunks of 80 lines
3. Process chunks CONCURRENTLY (parallel)
4. Each chunk:
   - Try batch translation with separator
   - If fails: Split into smaller batches (recursive)
   - If still fails: Translate individually (concurrent)
5. Merge results → Save to file + DB
```

### Batch Strategy

```
Original: 500 lines
         ↓
Split into 7 chunks of ~80 lines each
         ↓
Process ALL 7 chunks in PARALLEL
         ↓
Each chunk tries:
1. Single batch request (80 lines)
2. If fails: 2 batches of 40 lines
3. If fails: 40 concurrent individual requests
```

### Typical Performance

| Subtitle Size | Lines | Time (Go) |
|---------------|-------|-----------|
| Small | 100 | ~2-3s |
| Medium | 500 | ~8-12s |
| Large | 1000 | ~15-20s |

*Note: Actual time depends on Google Translate API response time*

### Monitoring

Check server logs to see translation progress:

```bash
2024/01/15 10:30:00 Starting translation of 450 text lines...
2024/01/15 10:30:02 Batch translation successful for chunk of 80 items
2024/01/15 10:30:03 Batch translation successful for chunk of 80 items
2024/01/15 10:30:04 Batch translation successful for chunk of 80 items
...
2024/01/15 10:30:10 Translation completed successfully
```
