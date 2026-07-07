# Technology Stack - Herlin AI Assistant

## Overview

Herlin AI Assistant is built with a modern, cloud-native technology stack designed for high performance, scalability, and maintainability.

## Backend Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway Layer                         │
│                    (Gin/Fiber Router)                        │
└─────────────────────────────────────────────────────────────┘
                           │
    ┌──────────────────────┼──────────────────────┐
    │                      │                      │
    ▼                      ▼                      ▼
┌──────────┐      ┌──────────────┐       ┌──────────────┐
│   Auth   │      │     AI       │       │    Files     │
└──────────┘      └──────────────┘       └──────────────┘
    │                      │                      │
    ▼                      ▼                      ▼
┌──────────┐      ┌──────────────┐       ┌──────────────┐
│PostgreSQL │      │   OpenAI     │       │    MinIO     │
│  (Users) │      │  (LLM APIs)  │       │ (File Store) │
└──────────┘      └──────────────┘       └──────────────┘
                           │
                           ▼
    ┌───────────────────────┬───────────────────────┐
    │                       │                       │
    ▼                       ▼                       ▼
┌──────────┐      ┌──────────────┐       ┌──────────────┐
│  Redis   │      │   Qdrant     │       │    SMTP      │
│  (Cache) │      │ (Vectors)    │       │ (Email)      │
└──────────┘      └──────────────┘       └──────────────┘
```

## Languages & Frameworks

| Component | Technology | Version | Purpose |
|-----------|------------|---------|---------|
| Backend | Go | 1.21+ | Core API server |
| Web Framework | Gin | v1.9.1 | HTTP routing |
| ORM | GORM | v1.25.5 | Database operations |
| Real-time | Gorilla WebSocket | v1.5.3 | WebSocket connections |

## Databases

### Primary Database: PostgreSQL
- **Purpose:** User data, conversations, messages, agents
- **Features:** ACID compliance, JSONB support, full-text search
- **Migrations:** Auto-migration via GORM

### Cache: Redis
- **Purpose:** Session caching, API response caching
- **Features:** In-memory key-value store, pub/sub, TTL
- **Integration:** go-redis/v9 client

### Vector Database: Qdrant/ChromaDB
- **Purpose:** AI memory embeddings, semantic search
- **Features:** Similarity search, metadata filtering
- **Embedding Model:** OpenAI text-embedding-ada-002

## AI Services Integration

### Large Language Models

| Provider | Model | Use Cases |
|----------|-------|-----------|
| OpenAI | GPT-4, GPT-3.5-turbo | General chat, code, vision |
| Google | Gemini Pro | Alternative chat provider |
| Anthropic | Claude 3 | High-quality responses |

### Specialized AI Services

| Service | Technology | Use Cases |
|---------|------------|-----------|
| Whisper | OpenAI API | Speech-to-text |
| TTS | OpenAI TTS-1 | Text-to-speech |
| Vision | GPT-4 Vision | Image analysis |
| Embeddings | text-embedding-ada-002 | Vector embeddings |

## Cloud Services

### Storage
| Service | Use Cases |
|---------|-----------|
| MinIO | S3-compatible local storage |
| AWS S3 | Cloud file storage |

### Infrastructure
| Service | Purpose |
|---------|---------|
| Docker | Containerization |
| Docker Compose | Local development |
| Kubernetes | Production orchestration |
| GitHub Actions | CI/CD |

## Development Tools

| Tool | Purpose |
|------|---------|
| Git | Version control |
| Go modules | Dependency management |
| Docker | Containerization |
| VSCode | IDE |

## Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=herlin
DB_PASSWORD=herlin123
DB_NAME=herlin_db
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=168h

# AI Providers
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=gpt-4
GEMINI_API_KEY=your-api-key
GEMINI_MODEL=gemini-pro
CLAUDE_API_KEY=your-api-key
CLAUDE_MODEL=claude-3-opus-20240229
AI_DEFAULT_PROVIDER=openai

# Storage
STORAGE_TYPE=minio
STORAGE_ENDPOINT=localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_BUCKET=herlin-files

# Vector DB
QDRANT_HOST=localhost
QDRANT_PORT=6333

# Email
EMAIL_USERNAME=
EMAIL_PASSWORD=
```

## API Rate Limits

| Endpoint | Limit | Notes |
|----------|-------|-------|
| Chat | 60/min | Per user |
| Document Upload | 10/min | File size < 10MB |
| Voice Transcription | 60/min | Per user |
| Image Analysis | 30/min | Per user |

## Security Considerations

- JWT tokens for authentication
- Bcrypt password hashing
- CORS enabled for cross-origin
- SQL injection prevention via GORM
- Rate limiting to be implemented
- HTTPS in production