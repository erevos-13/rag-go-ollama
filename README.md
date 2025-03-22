
# RAG-Go-Ollama

A Retrieval-Augmented Generation (RAG) system built with Go that uses Ollama for embeddings and LLM capabilities, with Chroma as the vector database for document storage and retrieval.

## Architecture Overview

This project implements a document Q&A system with the following components:

- **Go Backend**: Uses Gin framework for HTTP routing
- **Chroma DB**: Vector database running in a Docker container
- **Ollama**: LLM and embedding model running locally on the host machine
- **Docker**: Containerizes the application while connecting to local Ollama

## Features

- Upload PDF documents for processing and indexing
- Split documents into manageable chunks for embedding
- Generate embeddings using Ollama's embedding models
- Store document chunks and embeddings in Chroma vector database
- Search for relevant information using semantic similarity
- Generate comprehensive answers to questions using retrieved context

## Prerequisites

- **Docker** and **Docker Compose**
- **Ollama** running locally on your host machine
- Go 1.23+ (only needed for development)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/rag-go-ollama.git
   cd rag-go-ollama
   ```

2. Make sure Ollama is running locally with the required models:
   ```bash
   # Install models if you haven't already
   ollama pull gemma3:1b
   ollama pull nomic-embed-text
   
   # Ensure Ollama service is running
   ollama serve
   ```

3. Start the application using Docker Compose:
   ```bash
   docker compose up -d
   ```

## Configuration

The application uses environment variables for configuration, which are defined in the `docker-compose.yml` file:

- **CHROMA_URL**: URL to connect to the Chroma database (`http://chroma:8000`)
- **OLLAMA_MODEL**: Model to use for text generation (`gemma3:1b`)
- **OLLAMA_EMBEDDING_MODEL**: Model to use for generating embeddings (`nomic-embed-text`)
- **OLLAMA_HOST**: Host for Ollama service (`host.docker.internal`)
- **OLLAMA_URL**: URL for Ollama API (`http://host.docker.internal:11434`)

## API Endpoints

### Upload a Document
```
POST /document
```
Form parameters:
- `id`: Document identifier
- `title`: Document title (optional)
- `file`: PDF file to upload

### Search for Information
```
POST /document/search
```
JSON payload:
```json
{
  "query": "Your question about documents here"
}
```

## Technical Details

### Connection Management
The project uses a singleton connection manager (`conn_manager.go`) to handle connections to Ollama and Chroma services. This ensures efficient resource usage and provides thread-safe access to these services.

### Document Processing
Documents go through the following pipeline:
1. PDF loading using langchainGo's document loader
2. Text splitting into chunks of approximately 1000 characters with 100 character overlap
3. Embedding generation using Ollama's embedding model
4. Storage in Chroma DB with metadata about the source document

### Retrieval and Generation
When a query is received:
1. The query is embedded using the same embedding model
2. Similar documents are retrieved from Chroma using cosine similarity
3. Retrieved documents are used as context for the LLM
4. The LLM generates a comprehensive answer based on the provided context

## Development

To develop or modify the application:

1. Install Go 1.23+
2. Clone the repository
3. Run `go mod download` to install dependencies
4. Make your changes
5. Build with `go build -o main .`
6. Or use Docker to build: `docker build -t rag-go-ollama .`

## Evaluation

This project demonstrates a well-architected RAG system with:

- **Efficient Resource Management**: Using connection pooling and singletons
- **Docker Integration**: Running the application in containers while connecting to host services
- **Separation of Concerns**: Clear separation between routes, models, and connection management
- **Thread Safety**: Proper mutex usage for concurrent operations
- **Configuration Flexibility**: Environment variable-based configuration
- **Error Handling**: Comprehensive error propagation

Potential improvements could include:
- Adding authentication/authorization
- Supporting more document formats
- Implementing caching for frequently asked questions
- Adding logging and monitoring
- Creating a user interface
