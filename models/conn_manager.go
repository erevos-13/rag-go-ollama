package models

import (
	"fmt"
	"os"
	"sync"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

type ConnectionManager struct {
	// Ollama connection management
	ollamaClientOnce sync.Once
	ollamaClient     *ollama.LLM
	ollamaClientMu   sync.RWMutex

	// Chroma connection management
	chromaClientOnce sync.Once
	chromaClient     *chroma.Store
	chromaClientMu   sync.RWMutex

	// Embedder instance (depends on Ollama client)
	embedder     embeddings.Embedder
	embedderOnce sync.Once
	embedderMu   sync.RWMutex

	// LLM client management
	ollmaLLMClient     *ollama.LLM
	ollmaLLMClientOnce sync.Once
	ollmaLLMClientMu   sync.RWMutex
}

var (
	instance     *ConnectionManager
	instanceOnce sync.Once
)

func GetConnectionManager() *ConnectionManager {
	instanceOnce.Do(func() {
		instance = &ConnectionManager{}
	})
	return instance
}

func (cm *ConnectionManager) GetOllamaClient() (*ollama.LLM, error) {
	cm.ollamaClientMu.RLock()
	if cm.ollamaClient != nil {
		defer cm.ollamaClientMu.RUnlock()
		return cm.ollamaClient, nil
	}
	cm.ollamaClientMu.RUnlock()

	cm.ollamaClientMu.Lock()
	defer cm.ollamaClientMu.Unlock()

	var initErr error
	cm.ollamaClientOnce.Do(func() {
		model := os.Getenv("OLLAMA_EMBEDDING_MODEL")
		if model == "" {
			model = "all-minilm"
		}

		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}

		client, err := ollama.New(
			ollama.WithServerURL(ollamaURL),
			ollama.WithModel(model),
		)
		if err != nil {
			initErr = err
			return
		}
		cm.ollamaClient = client
		fmt.Println("Initialized shared Ollama client")
	})

	return cm.ollamaClient, initErr
}

func (cm *ConnectionManager) GetEmbedder() (embeddings.Embedder, error) {
	cm.embedderMu.RLock()
	if cm.embedder != nil {
		defer cm.embedderMu.RUnlock()
		return cm.embedder, nil
	}
	cm.embedderMu.RUnlock()

	cm.embedderMu.Lock()
	defer cm.embedderMu.Unlock()

	var initErr error
	cm.embedderOnce.Do(func() {
		client, err := cm.GetOllamaClient()
		if err != nil {
			initErr = err
			return
		}

		embedder, err := embeddings.NewEmbedder(client)
		if err != nil {
			initErr = err
			return
		}
		cm.embedder = embedder
		fmt.Println("Initialized shared embedder")
	})

	return cm.embedder, initErr
}

// GetChromaClient returns a shared Chroma client instance
func (cm *ConnectionManager) GetChromaClient() (*chroma.Store, error) {
	cm.chromaClientMu.RLock()
	if cm.chromaClient != nil {
		defer cm.chromaClientMu.RUnlock()
		return cm.chromaClient, nil
	}
	cm.chromaClientMu.RUnlock()

	cm.chromaClientMu.Lock()
	defer cm.chromaClientMu.Unlock()

	var initErr error
	cm.chromaClientOnce.Do(func() {
		embedder, err := cm.GetEmbedder()
		if err != nil {
			initErr = err
			return
		}

		chromaURL := os.Getenv("CHROMA_URL")
		if chromaURL == "" {
			chromaURL = "http://chroma:8000"
		}
		fmt.Println("chroma URL", chromaURL)

		store, err := chroma.New(
			chroma.WithChromaURL(chromaURL),
			chroma.WithDistanceFunction("cosine"),
			chroma.WithEmbedder(embedder),
		)
		if err != nil {
			initErr = err
			return
		}
		cm.chromaClient = &store
		fmt.Println("Initialized shared Chroma client")
	})

	return cm.chromaClient, initErr
}

// GetOllamaLLMClient returns a shared Ollama LLM client for text generation
func (cm *ConnectionManager) GetOllamaLLMClient() (*ollama.LLM, error) {
	cm.ollmaLLMClientMu.RLock()
	if cm.ollmaLLMClient != nil {
		defer cm.ollmaLLMClientMu.RUnlock()
		return cm.ollmaLLMClient, nil
	}
	cm.ollmaLLMClientMu.RUnlock()

	cm.ollmaLLMClientMu.Lock()
	defer cm.ollmaLLMClientMu.Unlock()

	var initErr error
	cm.ollmaLLMClientOnce.Do(func() {
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "gemma3:1b" // Default model for text generation
		}

		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}

		client, err := ollama.New(
			ollama.WithServerURL(ollamaURL),
			ollama.WithModel(model),
		)
		if err != nil {
			initErr = err
			return
		}
		cm.ollmaLLMClient = client
		fmt.Println("Initialized shared Ollama LLM client")
	})

	return cm.ollmaLLMClient, initErr
}
