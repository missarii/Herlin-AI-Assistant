# Herlin AI Assistant 🤖

### AI-Powered Personal Assistant Backend Platform (Built with Golang)

## Overview

Herlin AI Assistant is a modern AI-powered personal assistant platform that helps users communicate with an intelligent AI system, manage knowledge, analyze files, automate tasks, and receive personalized assistance.

The backend is built using **Golang** because Go excels at:
- High-performance APIs
- Concurrent processing
- Real-time communication
- Cloud-native applications
- Microservice architecture

## Tech Stack

### Backend
- **Golang** - Main programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for PostgreSQL
- **WebSocket** - Real-time communication with gorilla/websocket
- **gRPC** - For future microservice communication

### Database
- **PostgreSQL** - Main database for structured data
- **Redis** - Caching and session management
- **Qdrant/ChromaDB** - Vector database for AI memory

### AI Layer
- **OpenAI GPT** - Primary LLM provider
- **Google Gemini** - Alternative LLM provider
- **Anthropic Claude** - Additional LLM provider
- **Whisper** - Speech-to-text for voice features
- **GPT-4 Vision** - Image analysis capabilities

### Storage
- **MinIO** - S3-compatible file storage
- **AWS S3** - Cloud storage option

## Features

### 1. AI Chat System ⭐⭐⭐⭐⭐
- Text conversations with context understanding
- Multiple chat sessions
- AI-generated responses
- Markdown support
- Code explanation
- Translation capabilities

### 2. AI Model Integration
- **Cloud Models:** OpenAI GPT, Google Gemini, Anthropic Claude
- **Local Models:** Llama, Mistral, Gemma (planned)

### 3. Personal Memory System 🧠
- Long-term memory storage
- Semantic search with vector embeddings
- Importance ranking
- Access count tracking

### 4. Document Intelligence 📄
- PDF, Word, text file support
- Summarization
- Question answering
- Document search
- Information extraction

### 5. Voice Assistant 🎤
- Speech-to-text (Whisper)
- Text-to-speech
- Voice commands

### 6. AI Agent System ⭐⭐⭐⭐⭐
- Autonomous task execution
- Web search capabilities
- File operations
- API calls
- Data analysis

## Project Structure

```
Herlin-AI-Assistant/
├── backend/
│   ├── cmd/
│   │   └── server/
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── database/
│   │   └── database.go        # Database initialization
│   ├── internal/
│   │   ├── agents/            # AI agent system
│   │   ├── ai/                # AI service integration
│   │   ├── auth/              # Authentication & JWT
│   │   ├── cache/             # Redis caching
│   │   ├── chat/              # Chat system
│   │   ├── code/              # Code assistant
│   │   ├── files/             # Document processing
│   │   ├── memory/            # Memory system
│   │   ├── notifications/     # Notification system
│   │   ├── storage/           # MinIO/S3 storage
│   │   ├── users/             # User management
│   │   ├── vision/            # Image analysis
│   │   ├── voice/             # Voice assistant
│   │   └── websocket/         # Real-time communication
│   ├── middleware/
│   ├── models/                # Data models
│   ├── main.go                # Server entry point
│   ├── go.mod
│   └── go.sum
├── frontend/                  # Frontend application
├── mobile/                    # Mobile application
├── docker-compose.yml         # Docker configuration
├── Dockerfile               # Backend Docker image
└── .env.example             # Environment variables
```

## API Endpoints with Functions and Usage

### Authentication
- `POST /api/v1/public/register` - **Register new user account**
  ```json
  // Request body: { "name": "string", "email": "string", "password": "string" }
  ```
- `POST /api/v1/public/login` - **Authenticate user and get JWT token**
  ```json
  // Request body: { "email": "string", "password": "string" }
  // Returns: { "token", "refresh_token", "expires_at", "user" }
  ```
- `POST /api/v1/public/refresh` - **Get new JWT token using refresh token**
  ```json
  // Request body: { "refresh_token": "string" }
  ```

### User Management
- `GET /api/v1/users/me` - **Get current authenticated user profile**
  ```json
  // Headers: Authorization: Bearer <token>
  // Returns: { "id", "name", "email", "avatar", "created_at" }
  ```
- `PUT /api/v1/users/me` - **Update user profile information**
  ```json
  // Request body: { "name": "string", "email": "string" }
  ```

### Chat System
- `POST /api/v1/chat` - **Send message to AI and get response**
  ```json
  // Headers: Authorization: Bearer <token>
  // Request body: { "message": "string", "conversation_id": "uint?", "model": "string?" }
  // Returns: { "message_id", "conversation_id", "role", "content", "tokens", "created_at" }
  ```
- `POST /api/v1/chat/stream` - **Stream AI response in real-time (SSE)**
  ```json
  // Same as above but returns Server-Sent Events
  // Response: data: { "content": "string", "done": "bool" }
  ```

### Conversations
- `GET /api/v1/conversations` - **Get all user conversations**
  ```json
  // Returns array of: { "id", "title", "model", "is_archived", "created_at", "updated_at" }
  ```
- `POST /api/v1/conversations` - **Create new conversation**
  ```json
  // Request body: { "title": "string?", "model": "string?" }
  ```
- `GET /api/v1/conversations/:id` - **Get specific conversation details**
- `PUT /api/v1/conversations/:id` - **Update conversation**
  ```json
  // Request body: { "title": "string?", "model": "string?" }
  ```
- `DELETE /api/v1/conversations/:id` - **Delete conversation**
- `GET /api/v1/conversations/:id/messages` - **Get all messages in conversation**

### Code Assistant
- `POST /api/v1/code/generate` - **Generate code in specified language**
  ```json
  // Request body: { "code": "string?", "language": "string", "query": "string" }
  // Returns: { "generated_code", "explanation" }
  ```
- `POST /api/v1/code/debug` - **Debug code and identify errors**
  ```json
  // Request body: { "code": "string", "language": "string", "query": "error description" }
  // Returns: { "errors": [], "suggestions": [], "generated_code": "fixed code" }
  ```
- `POST /api/v1/code/review` - **Review code quality**
  ```json
  // Request body: { "code": "string", "language": "string" }
  // Returns: { "score": 1-10, "issues": [], "suggestions": [], "best_practices": [], "security": [] }
  ```
- `POST /api/v1/code/optimize` - **Optimize code performance**
  ```json
  // Request body: { "code": "string", "language": "string" }
  // Returns: { "generated_code", "optimizations": [], "explanation": "string" }
  ```
- `POST /api/v1/code/document` - **Generate documentation for code**
  ```json
  // Request body: { "code": "string", "language": "string" }
  // Returns: { "documentation": "string" }
  ```
- `POST /api/v1/code/explain` - **Explain code in simple terms**
  ```json
  // Request body: { "code": "string", "language": "string" }
  // Returns: { "explanation": "educational explanation" }
  ```
- `POST /api/v1/code/convert` - **Convert code between languages**
  ```json
  // Request body: { "code": "string", "source_language": "string", "query": "convert to X" }
  // Returns: { "generated_code": "converted code" }
  ```
- `POST /api/v1/code/tests` - **Generate unit tests**
  ```json
  // Request body: { "code": "string", "language": "string" }
  // Returns: { "generated_code": "test code" }
  ```
- `GET /api/v1/code/languages` - **Get supported programming languages**

### Documents
- `POST /api/v1/documents` - **Upload and process document**
  ```json
  // Form data: file (multipart)
  // Returns: { "id", "title", "file_name", "file_size", "file_type", "status": "processing|completed", "created_at" }
  ```
- `GET /api/v1/documents` - **Get all user documents**
- `POST /api/v1/documents/:id/analyze` - **Analyze document content**
  ```json
  // Returns: { "summary", "key_points": [], "topics": [], "language", "word_count", "page_count" }
  ```
- `POST /api/v1/documents/:id/query` - **Ask questions about document**
  ```json
  // Request body: { "query": "question about document" }
  // Returns: { "answer": "string" }
  ```
- `DELETE /api/v1/documents/:id` - **Delete document**

### Memory System
- `POST /api/v1/memory` - **Store memory with vector embedding**
  ```json
  // Request body: { "content": "string", "importance": 0.0-1.0 }
  // Creates embedding and stores in both PostgreSQL and Qdrant
  ```
- `POST /api/v1/memory/search` - **Semantic search memories**
  ```json
  // Request body: { "query": "search text", "limit": "int?" }
  // Returns: { "results": [{ "content", "importance", "similarity" }] }
  ```
- `GET /api/v1/memory` - **Get all memories**
  ```json
  // Query param: limit (default: 20)
  // Returns memories sorted by importance and access count
  ```
- `DELETE /api/v1/memory/:id` - **Delete memory**
- `PUT /api/v1/memory/:id` - **Update memory importance**
  ```json
  // Request body: { "importance": 0.0-1.0 }
  ```

### Vision/Image Analysis
- `POST /api/v1/vision/analyze` - **Analyze images**
  ```json
  // Request body: { "image_data": "base64", "image_format": "png|jpg|jpeg|gif", "query": "optional question" }
  // Returns: { "description", "objects": [], "text": "extracted text", "confidence": 0.0-1.0 }
  ```
- `POST /api/v1/vision/screenshot` - **Analyze screenshots/errors**
  ```json
  // Request body: { "image_data": "base64", "image_format": "png|jpg" }
  // Returns: { "error_explanation", "suggested_fix", "code_snippet" }
  ```
- `POST /api/v1/vision/diagram` - **Analyze diagrams and charts**
  ```json
  // Request body: { "image_data": "base64", "image_format": "png|jpg" }
  // Returns: { "type", "components": [], "flow", "summary" }
  ```

### Voice Assistant
- `POST /api/v1/voice/transcribe` - **Convert speech to text**
  ```json
  // Request body: { "audio_data": "base64", "audio_format": "mp3|wav|ogg" }
  // Returns: { "text", "language", "duration" }
  ```
- `POST /api/v1/voice/synthesize` - **Convert text to speech**
  ```json
  // Request body: { "text": "string", "voice": "alloy|echo|fable|onyx|nova|shimmer?", "language": "string?" }
  // Returns: audio/mp3 binary data
  ```
- `POST /api/v1/voice/command` - **Process voice commands**
  ```json
  // Request body: { "audio_data": "base64", "audio_format": "mp3|wav|ogg" }
  // Returns: { "command": "greeting|help|query|...", "transcription", "text" }
  ```

### AI Agents
- `POST /api/v1/agents` - **Create autonomous AI agent**
  ```json
  // Request body: { "name": "string", "description": "string", "capabilities": ["web_search", "file_operations", ...], "settings": {} }
  // Returns: { "id", "name", "description", "is_active", "created_at" }
  ```
- `GET /api/v1/agents` - **Get all user agents**
- `GET /api/v1/agents/:id` - **Get specific agent**
- `PUT /api/v1/agents/:id` - **Update agent configuration**
- `DELETE /api/v1/agents/:id` - **Delete agent**
- `POST /api/v1/agents/:id/execute` - **Execute agent task**
  ```json
  // Request body: { "task": "what to do" }
  // Returns: { "id": "task_id", "type", "input", "status": "processing" }
  ```
- `GET /api/v1/agents/tasks/:taskId` - **Get task status**

### Notifications
- `GET /api/v1/notifications` - **Get all notifications**
  ```json
  // Query param: limit (default: 20)
  // Returns: array of notifications
  ```
- `GET /api/v1/notifications/unread` - **Get unread notifications only**
- `PATCH /api/v1/notifications/:id/read` - **Mark notification as read**
- `PATCH /api/v1/notifications/read-all` - **Mark all notifications as read**
- `DELETE /api/v1/notifications/:id` - **Delete notification**
- `POST /api/v1/reminders` - **Create scheduled reminder**
  ```json
  // Request body: { "title": "string", "message": "string", "scheduled_at": "RFC3339 datetime" }
  ```
- `POST /api/v1/notifications/push` - **Send push notification**
  ```json
  // Request body: { "title": "string", "message": "string" }
  ```

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL
- Redis
- MinIO (optional, for file storage)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/missarii/Herlin-AI-Assistant.git
cd Herlin-AI-Assistant/backend
```

2. Copy environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Install dependencies:
```bash
go mod tidy
```

4. Run the server:
```bash
go run main.go
```

### Docker Deployment

```bash
docker-compose up -d
```

## Development Roadmap

### Phase 1 (Foundation) ✅
- Golang API
- User authentication
- PostgreSQL integration
- Basic chat API

### Phase 2 (AI) ✅
- LLM integration
- Chat history
- Streaming responses
- Memory system

### Phase 3 (Advanced) ✅
- Document AI
- Voice assistant
- AI agents
- Image understanding

### Phase 4 (Production) 🚧
- Docker production setup
- Kubernetes deployment
- Cloud deployment
- Monitoring

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please read CONTRIBUTING.md for guidelines.