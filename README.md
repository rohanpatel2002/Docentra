# AI Document Assistant
A backend system that allows users to upload documents and query them intelligently using semantic search.

## Tech Stack
- **Backend:** Go (Chi, GORM)
- **Embedding Service:** Python (FastAPI, sentence-transformers)
- **Database:** PostgreSQL
- **Infrastructure:** Docker (for local DB)

## Local Setup

### 1. Database
```bash
docker-compose up -d
```
### 2. Backend (Go)
```bash
cd backend
go mod tidy
go run cmd/api/main.go
```
### 3. Embedding Service (Python)
```bash
cd embedding-service
source venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```