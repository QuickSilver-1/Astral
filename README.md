# Rizq Backend

Backend service with PostgreSQL, Weaviate, and Presidio, containerized using Docker.

## 🐳 Docker Compose Services

| Service                 | Description                                        |
| ----------------------- | -------------------------------------------------- |
| **postgres**            | PostgreSQL database for core data storage          |
| **weaviate**            | Vector database for semantic search capabilities   |
| **contextionary**       | Weaviate component for natural language processing |
| **presidio-analyzer**   | Microsoft's sensitive data analyzer                |
| **presidio-anonymizer** | Tool for anonymizing sensitive data                |
| **app**                 | Main backend application                           |

## 🚀 Quick Start Guide

### 1. Install Task

**MacOS:**

```bash
brew install go-task/tap/go-task
```

**Windows (via Scoop):**

```powershell
scoop install task
```

**Linux:**

```bash
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

### 2. Start all services

```bash
docker-compose up -d --build
```

### 3. Run database migrations

```bash
task migrate:up
```

### 4. Access the API

Once running, the API is available at:

```
http://localhost:8080
```

## 🛠 Useful Commands

```bash
# View backend logs
docker-compose logs -f app

# Roll back migrations
task migrate:down

# View available Task commands
task

# Rebuild specific container
docker-compose up -d --build app
```

## 📋 Requirements

- Docker
- Docker Compose
- Task
- Go (for development)
