# Herlin AI Assistant 🤖

### AI-Powered Personal Assistant Backend Platform (Built with Golang)

![Herlin AI Assistant](https://images.openai.com/static-rsc-4/uWwiR5120GMK_LNJd8MdGVYTk0dKTPxPNdDeahsN2rH50pJQ_p6TN2nSOXPwje5_X_jDBNkE-SHKLj5NLyjYsgFRGtV5P5rDXfLZDlbJgCmCHZ2jUsmD9rJGczyswMLaOXILI5x7HuSxcwzde9ft-kuRm1ODiBy4w6YTW_izN8D9d9WxM6IXay6AJZ_ydQ5J?purpose=fullsize)

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

## API Endpoints

### Authentication
- `POST /api/v1/public/register` - Register new user
- `POST /api/v1/public/login` - User login
- `POST /api/v1/public/refresh` - Refresh JWT token

### User Management
- `GET /api/v1/users/me` - Get current user
- `PUT /api/v1/users/me` - Update user profile

### Conversations
- `GET /api/v1/conversations` - Get all conversations
- `POST /api/v1/conversations` - Create conversation
- `GET /api/v1/conversations/:id` - Get conversation
- `PUT /api/v1/conversations/:id` - Update conversation
- `DELETE /api/v1/conversations/:id` - Delete conversation

### Chat
- `POST /api/v1/chat` - Send message
- `GET /api/v1/conversations/:id/messages` - Get messages

### Memory
- `POST /api/v1/memory` - Store memory
- `POST /api/v1/memory/search` - Search memories
- `GET /api/v1/memory` - Get all memories

### Agents
- `POST /api/v1/agents` - Create agent
- `GET /api/v1/agents` - Get all agents
- `POST /api/v1/agents/:id/execute` - Execute agent task

### Code
- `POST /api/v1/code/generate` - Generate code
- `POST /api/v1/code/debug` - Debug code
- `POST /api/v1/code/review` - Review code
- `POST /api/v1/code/optimize` - Optimize code
- `POST /api/v1/code/document` - Generate documentation

### Documents
- `POST /api/v1/documents` - Upload document
- `GET /api/v1/documents` - Get all documents
- `POST /api/v1/documents/:id/query` - Query document

### Notifications
- `GET /api/v1/notifications` - Get notifications
- `POST /api/v1/reminders` - Create reminder

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